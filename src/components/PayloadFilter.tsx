import { Form } from 'react-bootstrap'

type PayloadFilterProps = {
  minPayload: number
  onChange: (value: number) => void
}

export function PayloadFilter({ minPayload, onChange }: PayloadFilterProps) {
  return (
    <Form.Group className="filter-block">
      <Form.Label>Минимальная грузоподъёмность: 
        {minPayload.toLocaleString('ru-RU')} кг</Form.Label>
      <Form.Range min={0} max={65000} step={1000} value={minPayload}
        onChange={(event) => onChange(Number(event.target.value))} />
    </Form.Group>
  )
}
