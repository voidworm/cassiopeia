import { useEffect, useState } from 'react'
import { fetchInvestigators, incrementPlayCount, setPlayCount, type Investigator } from './api'
import { InvestigatorRow } from './InvestigatorRow'

type SortColumn = 'name' | 'playCount'
type SortDirection = 'asc' | 'desc'

function App() {
  const [investigators, setInvestigators] = useState<Investigator[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [adminMode, setAdminMode] = useState(false)
  const [sortColumn, setSortColumn] = useState<SortColumn>('playCount')
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc')

  function load() {
    setLoading(true)
    fetchInvestigators()
      .then((data) => {
        setInvestigators(data)
        setError(null)
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  function applyUpdate(updated: Investigator) {
    setInvestigators((prev) => prev.map((inv) => (inv.uid === updated.uid ? updated : inv)))
  }

  function handleSort(column: SortColumn) {
    if (sortColumn === column) {
      setSortDirection((d) => (d === 'asc' ? 'desc' : 'asc'))
    } else {
      setSortColumn(column)
      setSortDirection(column === 'name' ? 'asc' : 'desc')
    }
  }

  function sortMark(column: SortColumn) {
    if (sortColumn !== column) return null
    return <span className="sort-mark">{sortDirection === 'asc' ? '↑' : '↓'}</span>
  }

  const sortedInvestigators = [...investigators].sort((a, b) => {
    const result =
      sortColumn === 'name' ? a.name.localeCompare(b.name) : a.playCount - b.playCount
    return sortDirection === 'asc' ? result : -result
  })

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
              <th className="sortable" onClick={() => handleSort('name')}>
                Investigator{sortMark('name')}
              </th>
              <th className="sortable" onClick={() => handleSort('playCount')}>
                Play count{sortMark('playCount')}
              </th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {sortedInvestigators.map((inv) => (
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
