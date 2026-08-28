import { useEffect, useState } from 'react'
import type { Investigator } from './api'

type Props = {
  investigator: Investigator
  adminMode: boolean
  onIncrement: (uid: string) => void
  onSet: (uid: string, value: number) => void
}

export function InvestigatorRow({ investigator, adminMode, onIncrement, onSet }: Props) {
  const [draft, setDraft] = useState(String(investigator.playCount))

  useEffect(() => {
    if (adminMode) setDraft(String(investigator.playCount))
  }, [adminMode, investigator.playCount])

  function commit() {
    const parsed = Number(draft)
    if (!Number.isFinite(parsed) || parsed < 0) {
      setDraft(String(investigator.playCount))
      return
    }
    onSet(investigator.uid, Math.trunc(parsed))
  }

  return (
    <tr>
      <td>{investigator.name}</td>
      <td>{investigator.playCount}</td>
      <td className="actions">
        {adminMode ? (
          <input
            className="admin-input"
            type="number"
            min={0}
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            onBlur={commit}
            onKeyDown={(e) => e.key === 'Enter' && commit()}
          />
        ) : (
          <button onClick={() => onIncrement(investigator.uid)}>+1</button>
        )}
      </td>
    </tr>
  )
}
