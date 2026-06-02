import { apiGet, apiPost, apiPut, apiDelete } from "./client";

export type ScheduledTaskDTO = {
  id: string;
  name: string;
  cronExpr: string;
  action: string;
  paramsJson: string;
  enabled: boolean;
  lastRunAt: string | null;
  nextRunAt: string | null;
  createdAt: string;
  updatedAt: string;
};

export type ScheduledTaskCreateRequest = {
  name: string;
  cronExpr: string;
  action: string;
  paramsJson?: string;
  enabled?: boolean;
};

export type ScheduledTaskUpdateRequest = {
  name?: string;
  cronExpr?: string;
  action?: string;
  paramsJson?: string;
  enabled?: boolean;
};

export function listScheduledTasks(): Promise<ScheduledTaskDTO[]> {
  return apiGet<ScheduledTaskDTO[]>("/scheduled-tasks");
}

export function createScheduledTask(body: ScheduledTaskCreateRequest): Promise<ScheduledTaskDTO> {
  return apiPost<ScheduledTaskDTO>("/scheduled-tasks", body);
}

export function updateScheduledTask(id: string, body: ScheduledTaskUpdateRequest): Promise<ScheduledTaskDTO> {
  return apiPut<ScheduledTaskDTO>(`/scheduled-tasks/${id}`, body);
}

export function deleteScheduledTask(id: string): Promise<{ message: string }> {
  return apiDelete<{ message: string }>(`/scheduled-tasks/${id}`);
}
