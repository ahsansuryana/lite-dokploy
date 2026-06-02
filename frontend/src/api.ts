import type { Application, Deployment, DashboardItem } from './types'

const BASE = '/api'

async function request<T>(path: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...opts,
  })
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || res.statusText)
  }
  return res.json()
}

export const api = {
  dashboard: () => request<{ applications: DashboardItem[] }>('/dashboard'),

  listApps: () => request<Application[]>('/applications'),
  getApp: (id: string) => request<Application>(`/applications/${id}`),
  createApp: (data: Partial<Application>) =>
    request<Application>('/applications', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  updateApp: (id: string, data: Partial<Application>) =>
    request<Application>(`/applications/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  deleteApp: (id: string) =>
    fetch(`${BASE}/applications/${id}`, { method: 'DELETE' }),

  listDeployments: (appId: string) =>
    request<Deployment[]>(`/applications/${appId}/deployments`),

  deploy: (id: string) =>
    request<Deployment>(`/applications/${id}/deploy`, { method: 'POST' }),
  redeploy: (id: string) =>
    request<Deployment>(`/applications/${id}/redeploy`, { method: 'POST' }),
  restart: (id: string) =>
    request<{ status: string }>(`/applications/${id}/restart`, { method: 'POST' }),
  stop: (id: string) =>
    request<{ status: string }>(`/applications/${id}/stop`, { method: 'POST' }),
  start: (id: string) =>
    request<{ status: string }>(`/applications/${id}/start`, { method: 'POST' }),

  getDeploymentLog: (depId: string) =>
    fetch(`${BASE}/deployments/${depId}/logs`).then((r) => r.text()),

  getAppLogs: (appId: string) =>
    request<{ appId: string; logFile: string; preview: string }[]>(
      `/applications/${appId}/logs`,
    ),
}
