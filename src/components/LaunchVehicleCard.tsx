import { Card } from 'react-bootstrap'
import type { LaunchVehicle } from '../types/LaunchVehicle'
import defaultImage from '../assets/default-launch-vehicle.svg'

type LaunchVehicleCardProps = { vehicle: LaunchVehicle }

export function LaunchVehicleCard({ vehicle }: LaunchVehicleCardProps) {
  const imageSource = vehicle.imageUrl || defaultImage

  return (
    <Card className="vehicle-card h-100 shadow-sm">
      <Card.Img variant="top" src={imageSource} alt={vehicle.name} className="vehicle-card-image" />
      <Card.Body>
        <Card.Title>{vehicle.name}</Card.Title>
        <Card.Text>{vehicle.shortDescription}</Card.Text>
        <div className="vehicle-spec">Грузоподъёмность: <strong>{vehicle.payloadKg.toLocaleString('ru-RU')} кг</strong></div>
        <div className="vehicle-spec">Тяга у земли: <strong>{vehicle.seaLevelThrustKn.toLocaleString('ru-RU')} кН</strong></div>
        <div className="vehicle-likes">♥ {vehicle.likeCount}</div>
      </Card.Body>
    </Card>
  )
}
