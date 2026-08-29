import type { ClassInfo, Investigator } from './api'

type Props = {
  investigator: Investigator
  classesById: Map<number, ClassInfo>
}

export function InvestigatorRow({ investigator, classesById }: Props) {
  const cls = investigator.classId !== undefined ? classesById.get(investigator.classId) : undefined

  return (
    <tr>
      <td>
        {cls?.colour && (
          <span className="class-marker" style={{ backgroundColor: cls.colour }} title={cls.name} />
        )}
        {investigator.name}
      </td>
      <td>{investigator.playCount}</td>
    </tr>
  )
}
