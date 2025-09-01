# 🐳 Enhanced Oryx Docker Guide

คู่มือการใช้งาน Enhanced Oryx ผ่าน Docker ที่รองรับ features ใหม่ทั้งหมด

## ✨ Features ที่รองรับ

- **HLS Input**: รองรับการรับ HLS streams
- **SRT Input**: รองรับ SRT streams สูงสุด 2 streams ต่อ port
- **Bypass Transcoding**: ไม่ใช้ FFmpeg re-encoding
- **Advanced Monitoring**: ติดตาม bandwidth และ concurrent streams
- **Historical Data**: ดูข้อมูลย้อนหลังรายวัน/รายสัปดาห์/รายเดือน

## 🚀 Quick Start

### 1. Build และ Run ด้วย Docker Compose (แนะนำ)

```bash
# ให้สิทธิ์ script
chmod +x build-and-run.sh

# Run script
./build-and-run.sh

# หรือเลือก option 1 จาก menu
```

### 2. Build และ Run ด้วยคำสั่ง Docker

```bash
# Build image
docker build -f Dockerfile.enhanced -t enhanced-oryx:latest .

# Run ด้วย Docker Compose
docker-compose -f docker-compose.enhanced.yml up -d

# หรือ run standalone
docker run -d \
  --name enhanced-oryx \
  -p 2022:2022 -p 1935:1935 -p 8080:8080 \
  -p 1985:1985 -p 10080:10080 -p 80:80 \
  enhanced-oryx:latest
```

## 📁 Directory Structure

```
.
├── Dockerfile.enhanced          # Docker image สำหรับ Enhanced Oryx
├── docker-compose.enhanced.yml  # Docker Compose configuration
├── build-and-run.sh            # Script สำหรับ build และ run
├── .dockerignore               # ไฟล์ที่ ignore ในการ build
├── data/                       # ข้อมูล persistent
├── logs/                       # Log files
├── config/                     # Configuration files
├── hls/                        # HLS output files
└── monitoring/                 # Monitoring configurations
    ├── prometheus/
    └── grafana/
```

## 🌐 Ports ที่เปิดใช้งาน

| Port | Service | Description |
|------|---------|-------------|
| 2022 | Oryx HTTP API | API สำหรับจัดการ HLS, SRT, Bypass Transcode |
| 1935 | RTMP | RTMP input/output |
| 8080 | HLS/HTTP-FLV | HLS streaming และ HTTP-FLV |
| 1985 | SRS HTTP API | SRS statistics และ API |
| 10080 | SRT | SRT input (default port) |
| 80 | Nginx | Web interface และ HLS output |
| 6379 | Redis | Cache และ configuration |

## 🔧 Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_ADDR` | `localhost:6379` | Redis server address |
| `SRS_CONFIG` | `/app/config/srs.conf` | SRS configuration file |
| `ORYX_LOG_LEVEL` | `info` | Log level |
| `ORYX_ENABLE_HLS_INPUT` | `true` | Enable HLS input |
| `ORYX_ENABLE_SRT_INPUT` | `true` | Enable SRT input |
| `ORYX_ENABLE_BYPASS_TRANSCODE` | `true` | Enable bypass transcoding |
| `ORYX_ENABLE_MONITORING` | `true` | Enable monitoring |

## 📊 Monitoring

### Prometheus + Grafana

```bash
# Start monitoring stack
docker-compose -f docker-compose.enhanced.yml up -d prometheus grafana

# Access Grafana
# URL: http://localhost:3000
# Username: admin
# Password: admin
```

### Metrics ที่รองรับ

- **Bandwidth**: ปริมาณ bandwidth ที่ใช้
- **Concurrent Streams**: จำนวน streams ที่ active
- **Stream Types**: แยกตามประเภท (HLS, SRT, RTMP)
- **Performance**: Latency, throughput

## 🎯 Usage Examples

### 1. สร้าง HLS Input

```bash
curl -X POST http://localhost:2022/terraform/v1/hls/input/create \
  -H "Content-Type: application/json" \
  -d '{
    "token": "your_token",
    "name": "My HLS Stream",
    "url": "https://example.com/stream.m3u8",
    "enabled": true
  }'
```

### 2. สร้าง SRT Input

```bash
curl -X POST http://localhost:2022/terraform/v1/srt/input/create \
  -H "Content-Type: application/json" \
  -d '{
    "token": "your_token",
    "name": "My SRT Stream",
    "port": 10080,
    "enabled": true
  }'
```

### 3. สร้าง Bypass Transcode Task

```bash
curl -X POST http://localhost:2022/terraform/v1/bypass/transcode/create \
  -H "Content-Type: application/json" \
  -d '{
    "token": "your_token",
    "name": "Bypass Task",
    "inputType": "hls",
    "inputUrl": "https://input.com/stream.m3u8",
    "outputType": "rtmp",
    "outputUrl": "rtmp://localhost:1935/live/stream",
    "bypassMode": "passthrough"
  }'
```

### 4. ดู Monitoring Data

```bash
# Real-time metrics
curl -X POST http://localhost:2022/terraform/v1/monitoring/realtime \
  -H "Content-Type: application/json" \
  -d '{"token": "your_token"}'

# Historical data
curl -X POST http://localhost:2022/terraform/v1/monitoring/query \
  -H "Content-Type: application/json" \
  -d '{
    "token": "your_token",
    "type": "bandwidth",
    "period": "daily",
    "startTime": "2024-01-01T00:00:00Z",
    "endTime": "2024-01-31T23:59:59Z"
  }'
```

## 🛠️ Troubleshooting

### 1. Container ไม่ start

```bash
# ดู logs
docker logs enhanced-oryx

# ตรวจสอบ status
docker ps -a

# Restart container
docker restart enhanced-oryx
```

### 2. Port conflicts

```bash
# ตรวจสอบ ports ที่ใช้
netstat -tulpn | grep :2022

# เปลี่ยน ports ใน docker-compose.yml
```

### 3. Redis connection error

```bash
# ตรวจสอบ Redis
docker logs oryx-redis

# Restart Redis
docker restart oryx-redis
```

## 🔄 Maintenance

### Update Image

```bash
# Pull latest code
git pull

# Rebuild image
docker build -f Dockerfile.enhanced -t enhanced-oryx:latest .

# Restart containers
docker-compose -f docker-compose.enhanced.yml restart
```

### Backup Data

```bash
# Backup volumes
docker run --rm -v enhanced-oryx_data:/data -v $(pwd):/backup alpine tar czf /backup/oryx-data.tar.gz -C /data .

# Backup logs
docker cp enhanced-oryx:/app/logs ./backup-logs
```

### Cleanup

```bash
# Stop และ remove containers
docker-compose -f docker-compose.enhanced.yml down

# Remove images
docker rmi enhanced-oryx:latest

# Remove volumes
docker volume rm enhanced-oryx_redis-data
```

## 📚 Additional Resources

- [Enhanced Features Documentation](./ENHANCED_FEATURES.md)
- [SRS Configuration](./platform/containers/conf/srs.enhanced.conf)
- [API Documentation](./ENHANCED_FEATURES.md#api-endpoints)

## 🆘 Support

หากมีปัญหาหรือต้องการความช่วยเหลือ:

1. ตรวจสอบ logs: `docker logs enhanced-oryx`
2. ตรวจสอบ status: `docker ps -a`
3. ดู documentation ใน `ENHANCED_FEATURES.md`
4. ตรวจสอบ configuration files

---

**Enhanced Oryx** - Streaming Server ที่รองรับ HLS, SRT, Bypass Transcoding และ Advanced Monitoring 🎥📡
