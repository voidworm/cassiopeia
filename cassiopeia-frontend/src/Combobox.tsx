import { useEffect, useRef, useState } from 'react'

export type ComboboxOption = {
  id: number
  label: string
  group?: string
}

type Props = {
  id?: string
  options: ComboboxOption[]
  value: number | ''
  onChange: (id: number | '') => void
  placeholder?: string
}

function groupOptions(options: ComboboxOption[]) {
  const order: string[] = []
  const byGroup = new Map<string, ComboboxOption[]>()
  for (const opt of options) {
    const group = opt.group ?? ''
    if (!byGroup.has(group)) {
      order.push(group)
      byGroup.set(group, [])
    }
    byGroup.get(group)!.push(opt)
  }
  return order.map((group) => ({ group, items: byGroup.get(group)! }))
}

export function Combobox({ id, options, value, onChange, placeholder }: Props) {
  const selected = options.find((o) => o.id === value) ?? null
  const [text, setText] = useState(selected?.label ?? '')
  const [open, setOpen] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    setText(selected?.label ?? '')
  }, [selected?.label])

  const filtered = options.filter((o) => o.label.toLowerCase().includes(text.toLowerCase()))
  const groups = groupOptions(filtered)

  function selectOption(opt: ComboboxOption) {
    onChange(opt.id)
    setText(opt.label)
    setOpen(false)
    inputRef.current?.blur()
  }

  function handleBlur() {
    setOpen(false)
    setText(selected?.label ?? '')
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Escape') {
      setOpen(false)
      setText(selected?.label ?? '')
      inputRef.current?.blur()
    } else if (e.key === 'Enter') {
      e.preventDefault()
      if (filtered.length === 1) {
        selectOption(filtered[0])
      }
    }
  }

  return (
    <div className="combobox">
      <input
        id={id}
        ref={inputRef}
        className="combobox-input"
        type="text"
        placeholder={placeholder}
        value={text}
        onFocus={(e) => {
          setOpen(true)
          e.target.select()
        }}
        onChange={(e) => {
          setText(e.target.value)
          setOpen(true)
        }}
        onBlur={handleBlur}
        onKeyDown={handleKeyDown}
        autoComplete="off"
      />
      {open && (
        <div className="combobox-list">
          {groups.length === 0 && <div className="combobox-empty">No matches</div>}
          {groups.map(({ group, items }) => (
            <div key={group || '__ungrouped'} className="combobox-group">
              {group && <div className="combobox-group-label">{group}</div>}
              {items.map((opt) => (
                <div
                  key={opt.id}
                  className={`combobox-option${opt.id === value ? ' selected' : ''}`}
                  onMouseDown={(e) => e.preventDefault()}
                  onClick={() => selectOption(opt)}
                >
                  {opt.label}
                </div>
              ))}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
