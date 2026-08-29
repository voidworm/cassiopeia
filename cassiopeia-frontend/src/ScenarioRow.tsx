import type { Scenario } from './api'

type Props = {
  scenario: Scenario
}

export function ScenarioRow({ scenario }: Props) {
  return (
    <tr>
      <td>{scenario.name}</td>
      <td>{scenario.playCount}</td>
    </tr>
  )
}
