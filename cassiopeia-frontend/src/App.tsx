import { useEffect, useState } from 'react'
import { fetchInvestigators, incrementPlayCount, setPlayCount, type Investigator } from './api'
import { InvestigatorRow } from './InvestigatorRow'

function App() {
  const [investigators, setInvestigators] = useState<Investigator[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [adminMode, setAdminMode] = useState(false)

  function load() {
    setLoading(true)
    fetchInvestigators()
      .then((data) => {
        setInvestigators(data.sort((a, b) => b.playCount - a.playCount))
        setError(null)
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  function applyUpdate(updated: Investigator) {
    setInvestigators((prev) =>
      prev
        .map((inv) => (inv.uid === updated.uid ? updated : inv))
        .sort((a, b) => b.playCount - a.playCount),
    )
  }

  async function handleIncrement(uid: string) {
    applyUpdate(await incrementPlayCount(uid, 1))
  }

  async function handleSet(uid: string, value: number) {
    applyUpdate(await setPlayCount(uid, value))
  }

  return (
    <div className="page">
      <h1>arkham archivist</h1>

      {error && <p className="error">{error}</p>}
      {loading && <p>Loading…</p>}

      {!loading && !error && (
        <table className="investigators">
          <thead>
            <tr>
              <th>Investigator</th>
              <th>Play count</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {investigators.map((inv) => (
              <InvestigatorRow
                key={inv.uid}
                investigator={inv}
                adminMode={adminMode}
                onIncrement={handleIncrement}
                onSet={handleSet}
              />
            ))}
          </tbody>
        </table>
      )}

      {!loading && !error && (
        <button className="admin-toggle secondary" onClick={() => setAdminMode((v) => !v)}>
          {adminMode ? 'Done' : 'Admin'}
        </button>
      )}
    </div>
  )
}

export default App
