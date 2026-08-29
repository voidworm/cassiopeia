import { useEffect, useState } from 'react'
import {
  fetchCampaigns,
  fetchClasses,
  fetchInvestigators,
  fetchScenarios,
  type Campaign,
  type ClassInfo,
  type Investigator,
  type Scenario,
} from './api'
import { InvestigatorRow } from './InvestigatorRow'
import { ScenarioRow } from './ScenarioRow'
import { DetailsPage } from './DetailsPage'

type SortColumn = 'name' | 'playCount'
type SortDirection = 'asc' | 'desc'

function useHashView() {
  const [view, setView] = useState(() => (window.location.hash === '#details' ? 'details' : 'home'))

  useEffect(() => {
    function onHashChange() {
      setView(window.location.hash === '#details' ? 'details' : 'home')
    }
    window.addEventListener('hashchange', onHashChange)
    return () => window.removeEventListener('hashchange', onHashChange)
  }, [])

  return view
}

const PAGE_SIZE = 10

function App() {
  const view = useHashView()

  const [investigators, setInvestigators] = useState<Investigator[]>([])
  const [scenarios, setScenarios] = useState<Scenario[]>([])
  const [classes, setClasses] = useState<ClassInfo[]>([])
  const [campaigns, setCampaigns] = useState<Campaign[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [sortColumn, setSortColumn] = useState<SortColumn>('playCount')
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc')

  const [hiddenClasses, setHiddenClasses] = useState<Set<number>>(new Set())
  const [showAllInvestigators, setShowAllInvestigators] = useState(false)
  const [showAllScenarios, setShowAllScenarios] = useState(false)

  function toggleClass(classId: number) {
    setHiddenClasses((prev) => {
      const next = new Set(prev)
      if (next.has(classId)) {
        next.delete(classId)
      } else {
        next.add(classId)
      }
      return next
    })
  }

  function toggleAllClasses() {
    setHiddenClasses((prev) => (prev.size === 0 ? new Set(classes.map((c) => c.id)) : new Set()))
  }

  useEffect(() => {
    Promise.all([fetchInvestigators(), fetchScenarios(), fetchClasses(), fetchCampaigns()])
      .then(([inv, scen, cls, camp]) => {
        setInvestigators(inv)
        setScenarios(scen)
        setClasses(cls)
        setCampaigns(camp)
        setError(null)
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }, [])

  if (view === 'details') {
    return <DetailsPage classes={classes} campaigns={campaigns} scenarios={scenarios} />
  }

  const classesById = new Map(classes.map((c) => [c.id, c]))
  const campaignsById = new Map(campaigns.map((c) => [c.id, c]))

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

  const sortedInvestigators = investigators
    .filter((inv) => inv.classId === undefined || !hiddenClasses.has(inv.classId))
    .sort((a, b) => {
      const result =
        sortColumn === 'name' ? a.name.localeCompare(b.name) : a.playCount - b.playCount
      return sortDirection === 'asc' ? result : -result
    })

  const sortedScenarios = [...scenarios].sort((a, b) => b.playCount - a.playCount || a.name.localeCompare(b.name))

  const visibleInvestigators = showAllInvestigators ? sortedInvestigators : sortedInvestigators.slice(0, PAGE_SIZE)
  const visibleScenarios = showAllScenarios ? sortedScenarios : sortedScenarios.slice(0, PAGE_SIZE)

  return (
    <div className="page">
      <h1>arkham archivist</h1>

      {error && <p className="error">{error}</p>}
      {loading && <p>Loading…</p>}

      {!loading && !error && (
        <>
          <div className="class-filter">
            <button
              type="button"
              className={`bookmark bookmark-all${hiddenClasses.size === 0 ? ' active' : ''}`}
              onClick={toggleAllClasses}
            >
              All
            </button>
            {classes.map((c) => (
              <button
                key={c.id}
                type="button"
                className={`bookmark${hiddenClasses.has(c.id) ? ' hidden' : ''}`}
                style={{ backgroundColor: c.colour }}
                onClick={() => toggleClass(c.id)}
              >
                {c.name}
              </button>
            ))}
          </div>

          <table className="data-table">
            <thead>
              <tr>
                <th className="sortable" onClick={() => handleSort('name')}>
                  Investigator{sortMark('name')}
                </th>
                <th className="sortable" onClick={() => handleSort('playCount')}>
                  Play count{sortMark('playCount')}
                </th>
              </tr>
            </thead>
            <tbody>
              {visibleInvestigators.map((inv) => (
                <InvestigatorRow key={inv.id} investigator={inv} classesById={classesById} />
              ))}
            </tbody>
          </table>

          {sortedInvestigators.length > PAGE_SIZE && (
            <button type="button" className="show-more" onClick={() => setShowAllInvestigators((v) => !v)}>
              {showAllInvestigators ? 'Show fewer' : `Show all ${sortedInvestigators.length}`}
            </button>
          )}

          <div className="section">
            <h2>Scenarios</h2>
            <table className="data-table">
              <thead>
                <tr>
                  <th>Scenario</th>
                  <th>Play count</th>
                </tr>
              </thead>
              <tbody>
                {visibleScenarios.map((sc) => (
                  <ScenarioRow key={sc.id} scenario={sc} campaignsById={campaignsById} />
                ))}
              </tbody>
            </table>
            {sortedScenarios.length > PAGE_SIZE && (
              <button type="button" className="show-more" onClick={() => setShowAllScenarios((v) => !v)}>
                {showAllScenarios ? 'Show fewer' : `Show all ${sortedScenarios.length}`}
              </button>
            )}
          </div>

          <a className="nav-link" href="#details">
            View classes, campaigns &amp; scenarios →
          </a>
        </>
      )}
    </div>
  )
}

export default App
