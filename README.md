# Лабораторная работа №5 — базовый SPA на React

Тема: каталог ракет-носителей для вывода спутников на орбиту.

## Технологии
React, TypeScript, Vite, React-Bootstrap, React Router, Bootstrap.

В ЛР5 данные берутся только из `src/mocks/launchVehicles.ts`. Backend, Redux и Context не используются.

## Страницы
### Лента `/feed`
Одна ракета на экран: название, описание, видео, грузоподъёмность, тяга у земли, лайки. Кнопка «Следующая ракета» переключает mock-объект.

### Плитка `/launch-vehicles`
Карточки ракет и фильтр по минимальной грузоподъёмности.

### Добавление `/add`
Макет формы: название, изображение, видео, кнопка «Далее». В ЛР5 отправка отключена.

## Navbar
Снизу 4 пункта: Лента, Добавление, Плитка, Вход. Доступны только Лента и Плитка.

## Изображение и видео по умолчанию
Если `imageUrl` пустой, `LaunchVehicleCard` использует локальный `default-launch-vehicle.svg`.
Если `videoUrl` пустой, `LaunchVehicleVideo` показывает локальный fallback-блок с постером.

## Props
`LaunchVehicleCard`: `vehicle`.
`PayloadFilter`: `minPayload`, `onChange`.
`LaunchVehicleVideo`: `videoUrl`, `poster`.

## Хуки
`useState`: текущая карточка ленты и значение фильтра.
`useEffect`: обновление title при смене ракеты.
`useMemo`: вычисление отфильтрованного списка.

## Запуск
```bash
npm install
npm run dev
```
