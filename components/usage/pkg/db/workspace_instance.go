// Copyright (c) 2022 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"strings"
	"time"
)

type WorkspaceInstance struct {
	ID                 uuid.UUID      `gorm:"primary_key;column:id;type:char;size:36;" json:"id"`
	WorkspaceID        string         `gorm:"column:workspaceId;type:char;size:36;" json:"workspaceId"`
	Configuration      datatypes.JSON `gorm:"column:configuration;type:text;size:65535;" json:"configuration"`
	Region             string         `gorm:"column:region;type:varchar;size:255;" json:"region"`
	ImageBuildInfo     sql.NullString `gorm:"column:imageBuildInfo;type:text;size:65535;" json:"imageBuildInfo"`
	IdeURL             string         `gorm:"column:ideUrl;type:varchar;size:255;" json:"ideUrl"`
	WorkspaceBaseImage string         `gorm:"column:workspaceBaseImage;type:varchar;size:255;" json:"workspaceBaseImage"`
	WorkspaceImage     string         `gorm:"column:workspaceImage;type:varchar;size:255;" json:"workspaceImage"`
	UsageAttributionID AttributionID  `gorm:"column:usageAttributionId;type:varchar;size:60;" json:"usageAttributionId"`

	CreationTime VarcharTime `gorm:"column:creationTime;type:varchar;size:255;" json:"creationTime"`
	StartedTime  VarcharTime `gorm:"column:startedTime;type:varchar;size:255;" json:"startedTime"`
	DeployedTime VarcharTime `gorm:"column:deployedTime;type:varchar;size:255;" json:"deployedTime"`
	StoppedTime  VarcharTime `gorm:"column:stoppedTime;type:varchar;size:255;" json:"stoppedTime"`
	LastModified time.Time   `gorm:"column:_lastModified;type:timestamp;default:CURRENT_TIMESTAMP(6);" json:"_lastModified"`
	StoppingTime VarcharTime `gorm:"column:stoppingTime;type:varchar;size:255;" json:"stoppingTime"`

	LastHeartbeat string         `gorm:"column:lastHeartbeat;type:varchar;size:255;" json:"lastHeartbeat"`
	StatusOld     sql.NullString `gorm:"column:status_old;type:varchar;size:255;" json:"status_old"`
	Status        datatypes.JSON `gorm:"column:status;type:json;" json:"status"`
	// Phase is derived from Status by extracting JSON from it. Read-only (-> property).
	Phase          sql.NullString `gorm:"->:column:phase;type:char;size:32;" json:"phase"`
	PhasePersisted string         `gorm:"column:phasePersisted;type:char;size:32;" json:"phasePersisted"`

	// deleted is restricted for use by db-sync
	_ bool `gorm:"column:deleted;type:tinyint;default:0;" json:"deleted"`
}

// WorkspaceRuntimeSeconds computes how long this WorkspaceInstance has been running.
// If the instance is still running (no stop time set), maxStopTime is used to to compute the duration - this is an upper bound on stop
func (i *WorkspaceInstance) WorkspaceRuntimeSeconds(maxStopTime time.Time) uint64 {
	start := i.CreationTime.Time()
	stop := maxStopTime

	if i.StoppedTime.IsSet() {
		if i.StoppedTime.Time().Before(maxStopTime) {
			stop = i.StoppedTime.Time()
		}
	}

	return uint64(stop.Sub(start).Round(time.Second).Seconds())
}

// TableName sets the insert table name for this struct type
func (i *WorkspaceInstance) TableName() string {
	return "d_b_workspace_instance"
}

// ListWorkspaceInstancesInRange lists WorkspaceInstances between from (inclusive) and to (exclusive).
// This results in all instances which have existed in the specified period, regardless of their current status, this includes:
// - terminated
// - running
// - instances which only just terminated after the start period
// - instances which only just started in the period specified
func ListWorkspaceInstancesInRange(ctx context.Context, conn *gorm.DB, from, to time.Time) ([]WorkspaceInstance, error) {
	var instances []WorkspaceInstance
	var instancesInBatch []WorkspaceInstance
	tx := conn.WithContext(ctx).
		Where(
			conn.Where("stoppedTime >= ?", TimeToISO8601(from)).Or("stoppedTime = ?", ""),
		).
		Where("creationTime < ?", TimeToISO8601(to)).
		Where("startedTime != ?", "").
		Where("usageAttributionId != ?", "").
		FindInBatches(&instancesInBatch, 1000, func(_ *gorm.DB, _ int) error {
			instances = append(instances, instancesInBatch...)
			return nil
		})
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to list workspace instances: %w", tx.Error)
	}

	return instances, nil
}

const (
	AttributionEntity_User = "user"
	AttributionEntity_Team = "team"
)

func newAttributionID(entity, identifier string) AttributionID {
	return AttributionID(fmt.Sprintf("%s:%s", entity, identifier))
}

func NewUserAttributionID(userID string) AttributionID {
	return newAttributionID(AttributionEntity_User, userID)
}

func NewTeamAttributionID(teamID string) AttributionID {
	return newAttributionID(AttributionEntity_Team, teamID)
}

// AttributionID consists of an entity, and an identifier in the form:
// <entity>:<identifier>, e.g. team:a7dcf253-f05e-4dcf-9a47-cf8fccc74717
type AttributionID string

func (a AttributionID) Values() (entity string, identifier string) {
	tokens := strings.Split(string(a), ":")
	if len(tokens) != 2 {
		return "", ""
	}

	return tokens[0], tokens[1]
}
