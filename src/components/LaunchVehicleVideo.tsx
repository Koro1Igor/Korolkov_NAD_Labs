type LaunchVehicleVideoProps = {
  videoUrl: string
  poster: string
}

export function LaunchVehicleVideo({ videoUrl, poster }: LaunchVehicleVideoProps) {
  if (!videoUrl) {
    return (
      <div className="feed-video fallback-video" style={{ backgroundImage: `url(${poster})` }}>
        <div className="fallback-video-overlay">
          <div className="fs-3 fw-bold">Видео по умолчанию</div>
          <div>Для mock-объекта не задан videoUrl</div>
        </div>
      </div>
    )
  }

  return <video className="feed-video" controls muted playsInline poster={poster} src={videoUrl} />
}
