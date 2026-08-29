import type { Campaign, ClassInfo, Scenario } from './api'

type Props = {
  classes: ClassInfo[]
  campaigns: Campaign[]
  scenarios: Scenario[]
}

export function DetailsPage({ classes, campaigns, scenarios }: Props) {
  const campaignsById = new Map(campaigns.map((c) => [c.id, c]))

  return (
    <div className="page">
      <h1>details</h1>

      <div className="section">
        <h2>Classes</h2>
        <table className="data-table">
          <thead>
            <tr>
              <th>Class</th>
            </tr>
          </thead>
          <tbody>
            {classes.map((c) => (
              <tr key={c.id}>
                <td>
                  {c.colour && (
                    <span className="class-marker" style={{ backgroundColor: c.colour }} />
                  )}
                  {c.name}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="section">
        <h2>Campaigns</h2>
        <table className="data-table">
          <thead>
            <tr>
              <th>Campaign</th>
            </tr>
          </thead>
          <tbody>
            {campaigns.map((c) => (
              <tr key={c.id}>
                <td>{c.name}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="section">
        <h2>Scenarios</h2>
        <table className="data-table">
          <thead>
            <tr>
              <th>Scenario</th>
              <th>Campaign</th>
              <th>Play count</th>
            </tr>
          </thead>
          <tbody>
            {scenarios.map((s) => (
              <tr key={s.id}>
                <td>{s.name}</td>
                <td>{s.campaignId !== undefined ? campaignsById.get(s.campaignId)?.name ?? '—' : '—'}</td>
                <td>{s.playCount}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <a className="nav-link" href="#">
        ← Back
      </a>
    </div>
  )
}
