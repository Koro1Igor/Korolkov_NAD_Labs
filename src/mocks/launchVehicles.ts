import type { LaunchVehicle } from '../types/LaunchVehicle'

export const launchVehicles: LaunchVehicle[] = [
  { id: 1, name: 'Falcon 9', shortDescription: 'Falcon 9 — ракета-носитель SpaceX. Максимальная полезная нагрузка на геопереходную орбиту — 8 300 кг.', payloadKg: 8300, seaLevelThrustKn: 7607, imageUrl: 'src/assets/launch_vehicles/falcon9.jpeg', videoUrl: 'src/assets/launch_vehicles/falcon9_launch.mp4', likeCount: 12 },
  { id: 2, name: 'Falcon Heavy', shortDescription: 'тяжёлая ракета-носитель SpaceX. Максимальная полезная нагрузка на геопереходную орбиту — 26 700 кг.', payloadKg: 26700, seaLevelThrustKn: 22819, imageUrl: 'src/assets/launch_vehicles/falcon_heavy.jpeg', videoUrl: 'src/assets/launch_vehicles/falcon_heavy_launch.mp4', likeCount: 19 },
  { id: 3, name: 'Ariane 64', shortDescription: 'четырёхбустерная конфигурация Ariane 6. Максимальная полезная нагрузка на стандартную геопереходную орбиту — 11 500 кг.', payloadKg: 11500, seaLevelThrustKn: 13700, imageUrl: 'src/assets/launch_vehicles/ariane64.jpeg', videoUrl: 'src/assets/launch_vehicles/ariane64_launch.mp4', likeCount: 8 },
  { id: 4, name: 'Протон-М', shortDescription: 'Тяжёлая российская ракета-носитель.', payloadKg: 23000, seaLevelThrustKn: 10000, imageUrl: 'src/assets/launch_vehicles/proton_m.jpeg', videoUrl: '', likeCount: 6 },
  { id: 5, name: 'Союз-2.1б', shortDescription: 'Ракета-носитель семейства «Союз» для орбитальных запусков.', payloadKg: 8200, seaLevelThrustKn: 4140, imageUrl: '', videoUrl: '', likeCount: 10 }
]
