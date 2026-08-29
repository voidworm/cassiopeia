import type { Campaign, Scenario } from './api'

type Props = {
  scenario: Scenario
  campaignsById: Map<number, Campaign>
}

export function ScenarioRow({ scenario, campaignsById }: Props) {
  const campaign = scenario.campaignId !== undefined ? campaignsById.get(scenario.campaignId) : undefined

  return (
    <tr>
      <td>
        {scenario.name}
        {campaign && <div className="scenario-campaign">{campaign.name}</div>}
      </td>
      <td>{scenario.playCount}</td>
    </tr>
  )
}
