/**
 * Copyright (c) 2022 Gitpod GmbH. All rights reserved.
 * Licensed under the GNU Affero General Public License (AGPL).
 * See License-AGPL.txt in the project root for license information.
 */

import { Attributes, Client } from "./types";
import { User } from "configcat-common/lib/RolloutEvaluator";
import { IConfigCatClient } from "configcat-common/lib/ConfigCatClient";

export const PROJECT_ID_ATTRIBUTE = "project_id";
export const TEAM_ID_ATTRIBUTE = "team_id";
export const TEAM_IDS_ATTRIBUTE = "team_ids";
export const TEAM_NAME_ATTRIBUTE = "team_name";
export const TEAM_NAMES_ATTRIBUTE = "team_names";

export class ConfigCatClient implements Client {
    private client: IConfigCatClient;

    constructor(cc: IConfigCatClient) {
        this.client = cc;
    }

    getValueAsync<T>(experimentName: string, defaultValue: T, attributes: Attributes): Promise<T> {
        return this.client.getValueAsync(experimentName, defaultValue, attributesToUser(attributes));
    }

    dispose(): void {
        return this.client.dispose();
    }
}

export function attributesToUser(attributes: Attributes): User {
    const userId = attributes.userId || "";
    const email = attributes.email || "";

    const custom: { [key: string]: string } = {};
    if (attributes.projectId) {
        custom[PROJECT_ID_ATTRIBUTE] = attributes.projectId;
    }
    if (attributes.teamId) {
        custom[TEAM_ID_ATTRIBUTE] = attributes.teamId;
    }
    if (attributes.teamName) {
        custom[TEAM_NAME_ATTRIBUTE] = attributes.teamName;
    }
    if (attributes.teams) {
        custom[TEAM_NAMES_ATTRIBUTE] = attributes.teams.map((t) => t.name).join(",");
        custom[TEAM_IDS_ATTRIBUTE] = attributes.teams.map((t) => t.id).join(",");
    }

    return new User(userId, email, "", custom);
}
