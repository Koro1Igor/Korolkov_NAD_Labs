import { useEffect, useState } from 'react'
import { Button, Card, Container } from 'react-bootstrap'
import { launchVehicles } from '../mocks/launchVehicles'
import defaultImage from '../assets/default-launch-vehicle.svg'
import { LaunchVehicleVideo } from '../components/LaunchVehicleVideo'

export function FeedPage() {
  const [index, setIndex] = useState(0)
  const vehicle = launchVehicles[index]
  const poster = vehicle.imageUrl || defaultImage

  useEffect(() => {
    document.title = `${vehicle.name} — Orbit`
  }, [vehicle.name])

  function showNext() {
    setIndex((current) => (current + 1) % launchVehicles.length)
  }

  return (
    <Container className="feed-shell">
      <Card className="feed-card shadow">
        <LaunchVehicleVideo videoUrl={vehicle.videoUrl} poster={poster} />
        <Card.Body>
          <p className="eyebrow">ЛЕНТА</p>
          <Card.Title as="h1">{vehicle.name}</Card.Title>
          <Card.Text>{vehicle.shortDescription}</Card.Text>
          <div className="feed-stats">
            <div><span>Грузоподъёмность</span><strong>{vehicle.payloadKg.toLocaleString('ru-RU')} кг</strong></div>
            <div><span>Тяга у земли</span><strong>{vehicle.seaLevelThrustKn.toLocaleString('ru-RU')} кН</strong></div>
            <div><span>Лайки</span><strong>♥ {vehicle.likeCount}</strong></div>
          </div>
          <Button className="w-100 mt-3" onClick={showNext}>Следующая ракета</Button>
        </Card.Body>
      </Card>
    </Container>
  )
}
