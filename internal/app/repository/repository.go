package repository

import "fmt"

const minioBaseURL = "http://localhost:9000/service-media"

type Service struct {
	ID           int
	Name         string
	Status       string
	Latitude     float64
	MaxPayloadKg int
	Description  string
	ImageURL     string
	VideoURL     string
	Likes        []int
}

type Repository struct {
	services []Service
}

func NewRepository() (*Repository, error) {
	services := []Service{
		{
			ID:           1,
			Name:         "Falcon 9",
			Status:       "published",
			Latitude:     28.50,
			MaxPayloadKg: 8300,
			Description:  "SpaceX Falcon 9. Максимальная заявленная полезная нагрузка на GTO — 8 300 кг в полностью расходуемой конфигурации.",
			ImageURL:     minioBaseURL + "/images/falcon9.jpeg",
			VideoURL:     minioBaseURL + "/videos/falcon9_launch.mp4",
			Likes:        []int{2, 5, 7, 11, 18, 21, 35, 41},
		},
		{
			ID:           2,
			Name:         "Falcon Heavy",
			Status:       "published",
			Latitude:     28.60,
			MaxPayloadKg: 26700,
			Description:  "SpaceX Falcon Heavy. Заявленная максимальная полезная нагрузка на GTO — 26 700 кг.",
			ImageURL:     minioBaseURL + "/images/falcon_heavy.jpeg",
			VideoURL:     minioBaseURL + "/videos/falcon_heavy_launch.mp4",
			Likes:        []int{1, 2, 3, 4, 7, 8, 10, 12, 14, 17, 19, 22},
		},
		{
			ID:           3,
			Name:         "Ariane 64",
			Status:       "published",
			Latitude:     5.24,
			MaxPayloadKg: 11500,
			Description:  "Ariane 64 — четырёхбустерная конфигурация Ariane 6. Для стандартной GTO указывается нагрузка до 11 500 кг.",
			ImageURL:     minioBaseURL + "/images/ariane64.jpeg",
			VideoURL:     minioBaseURL + "/videos/ariane64_launch.mp4",
			Likes:        []int{3, 6, 9, 12, 15, 18},
		},
		{
			ID:           4,
			Name:         "Soyuz ms16",
			Status:       "draft",
			Latitude:     5.24,
			MaxPayloadKg: 4500,
			Description:  "Черновик услуги.",
			ImageURL:     minioBaseURL + "/images/soyuz_ms16.jpeg",
			VideoURL:     minioBaseURL + "/videos/soyuz_ms16_launch.mp4",
			Likes:        []int{},
		},
		{
			ID:           5,
			Name:         "Falcon 9 Archive",
			Status:       "deleted",
			Latitude:     34.74,
			MaxPayloadKg: 8300,
			Description:  "Удалённая услуга. Она остаётся в коллекции, но в интерфейсе не отображается.",
			ImageURL:     minioBaseURL + "/images/falcon_9_archive.jpeg",
			VideoURL:     minioBaseURL + "/videos/falcon_9_archive.mp4",
			Likes:        []int{1},
		},
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("коллекция услуг пуста")
	}

	return &Repository{services: services}, nil
}

func (r *Repository) GetPublishedServices() ([]Service, error) {
	result := make([]Service, 0)
	for _, service := range r.services {
		if service.Status == "published" {
			result = append(result, service)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("опубликованные услуги не найдены")
	}
	return result, nil
}

func (r *Repository) GetServiceByID(id int) (Service, error) {
	services, err := r.GetPublishedServices()
	if err != nil {
		return Service{}, err
	}
	for _, service := range services {
		if service.ID == id {
			return service, nil
		}
	}
	return Service{}, fmt.Errorf("услуга с id=%d не найдена", id)
}

func (r *Repository) GetFirstPublishedService() (Service, error) {
	services, err := r.GetPublishedServices()
	if err != nil {
		return Service{}, err
	}
	return services[0], nil
}

func (r *Repository) GetNextPublishedService(id int) (Service, error) {
	services, err := r.GetPublishedServices()
	if err != nil {
		return Service{}, err
	}
	for i, service := range services {
		if service.ID == id {
			return services[(i+1)%len(services)], nil
		}
	}
	return Service{}, fmt.Errorf("услуга с id=%d не найдена", id)
}

func (r *Repository) GetDraftService() (Service, error) {
	for _, service := range r.services {
		if service.Status == "draft" {
			return service, nil
		}
	}
	return Service{}, fmt.Errorf("черновик не найден")
}

func (r *Repository) GetFilteredServices(minPayload int) ([]Service, error) {
	services, err := r.GetPublishedServices()
	if err != nil {
		return nil, err
	}
	if minPayload <= 0 {
		return services, nil
	}

	result := make([]Service, 0)
	for _, service := range services {
		if service.MaxPayloadKg >= minPayload {
			result = append(result, service)
		}
	}
	return result, nil
}
