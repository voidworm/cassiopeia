import { useState } from 'react'
import { createSession, type Campaign, type ClassInfo, type Investigator, type Scenario } from './api'
import { Combobox, type ComboboxOption } from './Combobox'

const MAX_INVESTIGATORS = 4
const NO_CAMPAIGN_LABEL = 'Standalone'
const NO_CLASS_LABEL = 'Other'

type Props = {
  scenarios: Scenario[]
  investigators: Investigator[]
  campaigns: Campaign[]
  classes: ClassInfo[]
  onClose: () => void
  onCreated: () => void
}

function buildScenarioOptions(scenarios: Scenario[], campaigns: Campaign[]): ComboboxOption[] {
  const campaignNameById = new Map(campaigns.map((c) => [c.id, c.name]))
  return scenarios
    .map((sc) => ({
      id: sc.id,
      label: sc.name,
      group: (sc.campaignId !== undefined && campaignNameById.get(sc.campaignId)) || NO_CAMPAIGN_LABEL,
    }))
    .sort((a, b) => {
      if (a.group !== b.group) {
        if (a.group === NO_CAMPAIGN_LABEL) return 1
        if (b.group === NO_CAMPAIGN_LABEL) return -1
        return a.group.localeCompare(b.group)
      }
      return a.label.localeCompare(b.label)
    })
}

function buildInvestigatorOptions(investigators: Investigator[], classes: ClassInfo[]): ComboboxOption[] {
  const classNameById = new Map(classes.map((c) => [c.id, c.name]))
  return investigators
    .map((inv) => ({
      id: inv.id,
      label: inv.name,
      group: (inv.classId !== undefined && classNameById.get(inv.classId)) || NO_CLASS_LABEL,
    }))
    .sort((a, b) => {
      if (a.group !== b.group) {
        if (a.group === NO_CLASS_LABEL) return 1
        if (b.group === NO_CLASS_LABEL) return -1
        return a.group.localeCompare(b.group)
      }
      return a.label.localeCompare(b.label)
    })
}

export function NewSessionModal({ scenarios, investigators, campaigns, classes, onClose, onCreated }: Props) {
  const [scenarioId, setScenarioId] = useState<number | ''>('')
  const [investigatorSlots, setInvestigatorSlots] = useState<(number | '')[]>([''])
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const scenarioOptions = buildScenarioOptions(scenarios, campaigns)
  const investigatorOptions = buildInvestigatorOptions(investigators, classes)

  function updateSlot(index: number, id: number | '') {
    setInvestigatorSlots((prev) => prev.map((v, i) => (i === index ? id : v)))
  }

  function addSlot() {
    setInvestigatorSlots((prev) => (prev.length < MAX_INVESTIGATORS ? [...prev, ''] : prev))
  }

  function removeSlot(index: number) {
    setInvestigatorSlots((prev) => (prev.length > 1 ? prev.filter((_, i) => i !== index) : prev))
  }

  const selectedInvestigatorIds = investigatorSlots.filter((v): v is number => v !== '')
  const canSave = scenarioId !== '' && selectedInvestigatorIds.length >= 1

  async function handleSave() {
    if (!canSave || typeof scenarioId !== 'number') return
    setSaving(true)
    setError(null)
    try {
      await createSession(scenarioId, selectedInvestigatorIds)
      onCreated()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'failed to save session')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <h2>New play session</h2>

        <label className="modal-label" htmlFor="scenario-select">
          Scenario
        </label>
        <Combobox
          id="scenario-select"
          options={scenarioOptions}
          value={scenarioId}
          onChange={setScenarioId}
          placeholder="Type to search scenarios…"
        />

        <div className="modal-label">
          Investigators ({selectedInvestigatorIds.length}/{MAX_INVESTIGATORS})
        </div>
        <div className="investigator-slots">
          {investigatorSlots.map((slotValue, index) => {
            const takenElsewhere = new Set(
              investigatorSlots.filter((_, i) => i !== index).filter((v): v is number => v !== ''),
            )
            const availableOptions = investigatorOptions.filter((opt) => !takenElsewhere.has(opt.id))
            return (
              <div className="investigator-slot-row" key={index}>
                <Combobox
                  options={availableOptions}
                  value={slotValue}
                  onChange={(id) => updateSlot(index, id)}
                  placeholder="Type to search investigators…"
                />
                {investigatorSlots.length > 1 && (
                  <button
                    type="button"
                    className="slot-remove"
                    aria-label="Remove investigator"
                    onClick={() => removeSlot(index)}
                  >
                    ×
                  </button>
                )}
              </div>
            )
          })}
        </div>
        {investigatorSlots.length < MAX_INVESTIGATORS && (
          <button type="button" className="add-investigator" onClick={addSlot}>
            + Add investigator
          </button>
        )}

        {error && <p className="error">{error}</p>}

        <div className="modal-actions">
          <button type="button" className="modal-button modal-cancel" onClick={onClose} disabled={saving}>
            Cancel
          </button>
          <button type="button" className="modal-button modal-save" onClick={handleSave} disabled={!canSave || saving}>
            {saving ? 'Saving…' : 'Save'}
          </button>
        </div>
      </div>
    </div>
  )
}
