export type Investigator = {
  id: number
  name: string
  classId?: number
  playCount: number
}

export type Scenario = {
  id: number
  name: string
  campaignId?: number
  playCount: number
}

export type ClassInfo = {
  id: number
  name: string
  colour?: string
}

export type Campaign = {
  id: number
  name: string
}

const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

async function handle<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error ?? `request failed with status ${res.status}`)
  }
  return res.json()
}

export function fetchInvestigators(): Promise<Investigator[]> {
  return fetch(`${API_URL}/api/investigators`).then((res) => handle(res))
}

export function fetchScenarios(): Promise<Scenario[]> {
  return fetch(`${API_URL}/api/scenarios`).then((res) => handle(res))
}

export function fetchClasses(): Promise<ClassInfo[]> {
  return fetch(`${API_URL}/api/classes`).then((res) => handle(res))
}

export function fetchCampaigns(): Promise<Campaign[]> {
  return fetch(`${API_URL}/api/campaigns`).then((res) => handle(res))
}

export type Session = {
  id: number
  scenarioId?: number
  timestamp: string
}

export function createSession(scenarioId: number, investigatorIds: number[]): Promise<Session> {
  return fetch(`${API_URL}/api/sessions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ scenarioId, investigatorIds }),
  }).then((res) => handle(res))
}
