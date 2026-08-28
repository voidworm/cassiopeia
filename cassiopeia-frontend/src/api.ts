export type Investigator = {
  uid: string
  name: string
  playCount: number
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

export function incrementPlayCount(uid: string, by = 1): Promise<Investigator> {
  return fetch(`${API_URL}/api/investigators/${uid}/increment`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ by }),
  }).then((res) => handle(res))
}

export function setPlayCount(uid: string, value: number): Promise<Investigator> {
  return fetch(`${API_URL}/api/investigators/${uid}/set`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ value }),
  }).then((res) => handle(res))
}
