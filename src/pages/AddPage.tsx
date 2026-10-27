import { Button, Container, Form } from 'react-bootstrap'

export function AddPage() {
  return (
    <Container className="page-shell">
      <div className="page-heading">
        <p className="eyebrow">ДОБАВЛЕНИЕ</p>
        <h1>Новая ракета-носитель</h1>
        <p>Макет страницы. В ЛР5 отправка данных отключена.</p>
      </div>
      <Form className="mock-form">
        <Form.Group className="mb-3"><Form.Label>Название</Form.Label><Form.Control placeholder="Название ракеты-носителя" disabled /></Form.Group>
        <Form.Group className="mb-3"><Form.Label>Изображение</Form.Label><Form.Control type="file" disabled /></Form.Group>
        <Form.Group className="mb-3"><Form.Label>Видео</Form.Label><Form.Control type="file" disabled /></Form.Group>
        <Button disabled>Далее</Button>
      </Form>
    </Container>
  )
}
