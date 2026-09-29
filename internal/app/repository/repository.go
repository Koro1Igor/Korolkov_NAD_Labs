package repository

import "fmt"

const minioBaseURL = "http://localhost:9000/service-media"

type LaunchVehicle struct {
	ID               int
	Name             string
	Status           string
	PayloadKg        int
	SeaLevelThrustKN int
	Description      string
	ImageURL         string
	VideoURL         string
	Likes            []int
}

type Repository struct {
	launchVehicles []LaunchVehicle
}

func NewRepository() (*Repository, error) {
	launchVehicles := []LaunchVehicle{
		{
			ID:               1,
			Name:             "Falcon 9",
			Status:           "published",
			PayloadKg:        8300,
			SeaLevelThrustKN: 7607,
			Description:      "Falcon 9 — ракета-носитель SpaceX. Максимальная полезная нагрузка на геопереходную орбиту — 8 300 кг.",
			ImageURL:         minioBaseURL + "/images/falcon9.jpeg",
			VideoURL:         minioBaseURL + "/videos/falcon9_launch.mp4",
			Likes:            []int{2, 5, 7, 11, 18, 21, 35, 41},
		},
		{
			ID:               2,
			Name:             "Falcon Heavy",
			Status:           "published",
			PayloadKg:        26700,
			SeaLevelThrustKN: 22819,
			Description:      "Falcon Heavy — тяжёлая ракета-носитель SpaceX. Максимальная полезная нагрузка на геопереходную орбиту — 26 700 кг.",
			ImageURL:         minioBaseURL + "/images/falcon_heavy.jpeg",
			VideoURL:         minioBaseURL + "/videos/falcon_heavy_launch.mp4",
			Likes:            []int{1, 2, 3, 4, 7, 8, 10, 12, 14, 17, 19, 22},
		},
		{
			ID:               3,
			Name:             "Ariane 64",
			Status:           "published",
			PayloadKg:        11500,
			SeaLevelThrustKN: 1370,
			Description:      "Ariane 64 — четырёхбустерная конфигурация Ariane 6. Максимальная полезная нагрузка на стандартную геопереходную орбиту — 11 500 кг.",
			ImageURL:         minioBaseURL + "/images/ariane64.jpeg",
			VideoURL:         minioBaseURL + "/videos/ariane64_launch.mp4",
			Likes:            []int{3, 6, 9, 12, 15, 18},
		},
		{
			ID:               4,
			Name:             "Soyuz MS16",
			Status:           "draft",
			PayloadKg:        4500,
			SeaLevelThrustKN: 1370,
			Description:      "Союз МС-16 — пилотируемый космический корабль серии „Союз МС“, предназначенный для доставки экипажа на Международную космическую станцию и возвращения его на Землю.",
			ImageURL:         minioBaseURL + "/images/soyuz_ms16.jpeg",
			VideoURL:         minioBaseURL + "/videos/soyuz_ms16_launch.mp4",
			Likes:            []int{},
		},
		{
			ID:               5,
			Name:             "Souyz MS12",
			Status:           "deleted",
			PayloadKg:        8300,
			SeaLevelThrustKN: 7607,
			Description:      "Удалённая запись ракеты-носителя. Она остаётся в коллекции, но не отображается в интерфейсе.",
			ImageURL:         minioBaseURL + "/images/souyz_ms12_archive.jpeg",
			VideoURL:         minioBaseURL + "/videos/souyz_ms12_launch_archive.mp4",
			Likes:            []int{1},
		},
	}

	if len(launchVehicles) == 0 {
		return nil, fmt.Errorf("коллекция launch vehicles пуста")
	}

	return &Repository{launchVehicles: launchVehicles}, nil
}

func (r *Repository) GetPublishedLaunchVehicles() ([]LaunchVehicle, error) {
	result := make([]LaunchVehicle, 0)
	for _, launchVehicle := range r.launchVehicles {
		if launchVehicle.Status == "published" {
			result = append(result, launchVehicle)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("опубликованные launch vehicles не найдены")
	}
	return result, nil
}

func (r *Repository) GetLaunchVehicleByID(id int) (LaunchVehicle, error) {
	launchVehicles, err := r.GetPublishedLaunchVehicles()
	if err != nil {
		return LaunchVehicle{}, err
	}
	for _, launchVehicle := range launchVehicles {
		if launchVehicle.ID == id {
			return launchVehicle, nil
		}
	}
	return LaunchVehicle{}, fmt.Errorf("launch vehicle с id=%d не найден", id)
}

func (r *Repository) GetFirstPublishedLaunchVehicle() (LaunchVehicle, error) {
	launchVehicles, err := r.GetPublishedLaunchVehicles()
	if err != nil {
		return LaunchVehicle{}, err
	}
	return launchVehicles[0], nil
}

func (r *Repository) GetNextPublishedLaunchVehicle(id int) (LaunchVehicle, error) {
	launchVehicles, err := r.GetPublishedLaunchVehicles()
	if err != nil {
		return LaunchVehicle{}, err
	}
	for i, launchVehicle := range launchVehicles {
		if launchVehicle.ID == id {
			return launchVehicles[(i+1)%len(launchVehicles)], nil
		}
	}
	return LaunchVehicle{}, fmt.Errorf("launch vehicle с id=%d не найден", id)
}

func (r *Repository) GetDraftLaunchVehicle() (LaunchVehicle, error) {
	for _, launchVehicle := range r.launchVehicles {
		if launchVehicle.Status == "draft" {
			return launchVehicle, nil
		}
	}
	return LaunchVehicle{}, fmt.Errorf("черновик launch vehicle не найден")
}

func (r *Repository) GetFilteredLaunchVehicles(minPayload, maxPayload int) ([]LaunchVehicle, error) {
	launchVehicles, err := r.GetPublishedLaunchVehicles()
	if err != nil {
		return nil, err
	}

	result := make([]LaunchVehicle, 0)
	for _, launchVehicle := range launchVehicles {
		if launchVehicle.PayloadKg >= minPayload && launchVehicle.PayloadKg <= maxPayload {
			result = append(result, launchVehicle)
		}
	}
	return result, nil
}
