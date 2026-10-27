import { useMemo, useState } from 'react'
import { Col, Container, Row } from 'react-bootstrap'
import { LaunchVehicleCard } from '../components/LaunchVehicleCard'
import { PayloadFilter } from '../components/PayloadFilter'
import { launchVehicles } from '../mocks/launchVehicles'

export function TilePage() {
  const [minPayload, setMinPayload] = useState(0)
  const filteredVehicles = useMemo(
    () => launchVehicles.filter((vehicle) => vehicle.payloadKg >= minPayload),
    [minPayload],
  )

  return (
    <Container className="page-shell">
      <div className="page-heading">
        <p className="eyebrow">КАТАЛОГ</p>
        <h1>Ракеты-носители</h1>
        <p>Сравнение ракет по грузоподъёмности и тяге у земли.</p>
      </div>
      <PayloadFilter minPayload={minPayload} onChange={setMinPayload} />
      <Row className="g-4">
        {filteredVehicles.map((vehicle) => (
          <Col key={vehicle.id} xs={12} md={6} xl={4}>
            <LaunchVehicleCard vehicle={vehicle} />
          </Col>
        ))}
      </Row>
    </Container>
  )
}
