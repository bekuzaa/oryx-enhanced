# Enhanced Oryx Features

This document describes the enhanced features added to Oryx streaming platform.

## Overview

The enhanced Oryx platform now supports:
1. **HLS Input Support** - Accept HLS streams as input sources
2. **SRT Input Support** - Accept SRT streams (max 2 per port) without Stream ID
3. **Bypass Transcoding** - Stream processing without FFmpeg re-encoding
4. **Advanced Monitoring** - Real-time bandwidth and concurrent stream monitoring
5. **Data Filtering** - SCTE-35 and video metadata filtering capabilities

## 1. HLS Input Support

### Features
- Accept HLS streams from external sources
- Automatic playlist parsing and segment handling
- Stream health monitoring
- Multiple concurrent HLS inputs

### API Endpoints
- `POST /terraform/v1/hls/input/create` - Create new HLS input
- `POST /terraform/v1/hls/input/query` - Query HLS inputs
- `POST /terraform/v1/hls/input/update` - Update HLS input
- `POST /terraform/v1/hls/input/delete` - Delete HLS input

### Configuration
```json
{
  "name": "External HLS Stream",
  "url": "https://example.com/stream.m3u8",
  "enabled": true
}
```

## 2. SRT Input Support

### Features
- Accept SRT streams without requiring Stream ID
- Maximum 2 concurrent streams per port
- Low-latency streaming support
- Connection monitoring and management

### API Endpoints
- `POST /terraform/v1/srt/input/create` - Create new SRT input
- `POST /terraform/v1/srt/input/query` - Query SRT inputs
- `POST /terraform/v1/srt/input/update` - Update SRT input
- `POST /terraform/v1/srt/input/delete` - Delete SRT input
- `POST /terraform/v1/srt/stream/query` - Query SRT streams

### Configuration
```json
{
  "name": "SRT Input Port",
  "port": 10080,
  "enabled": true,
  "maxStreams": 2
}
```

## 3. Bypass Transcoding

### Features
- Process streams without FFmpeg re-encoding
- Support for HLS, SRT, and RTMP inputs/outputs
- Configurable filtering modes
- SCTE-35 and video metadata handling

### API Endpoints
- `POST /terraform/v1/bypass/transcode/create` - Create bypass transcode task
- `POST /terraform/v1/bypass/transcode/query` - Query bypass transcode tasks
- `POST /terraform/v1/bypass/transcode/update` - Update bypass transcode task
- `POST /terraform/v1/bypass/transcode/delete` - Delete bypass transcode task

### Configuration
```json
{
  "name": "Bypass Task",
  "inputType": "hls",
  "inputUrl": "https://input.com/stream.m3u8",
  "outputType": "rtmp",
  "outputUrl": "rtmp://output.com/live/stream",
  "bypassMode": "passthrough",
  "enabled": true
}
```

### Bypass Modes
- **Passthrough**: Stream data passes through unchanged
- **Filter**: Apply configured filters to remove specific data types

## 4. Advanced Monitoring

### Features
- Real-time bandwidth monitoring
- Concurrent stream counting
- Historical data aggregation (daily, weekly, monthly)
- Performance metrics collection

### API Endpoints
- `POST /terraform/v1/monitoring/realtime` - Get real-time metrics
- `POST /terraform/v1/monitoring/query` - Query historical data
- `POST /terraform/v1/monitoring/config/query` - Get monitoring configuration
- `POST /terraform/v1/monitoring/config/update` - Update monitoring configuration

### Metrics Types
- **Bandwidth**: Network bandwidth usage in Mbps
- **Concurrent Streams**: Number of active streams
- **Stream Health**: Stream status and performance indicators

### Data Retention
- Configurable retention period (default: 30 days)
- Automatic data cleanup
- Redis-based storage with TTL

## 5. SRS Configuration Enhancements

### New Configuration File
The enhanced SRS configuration is located at:
```
platform/containers/conf/srs.enhanced.conf
```

### Key Features
- HLS input vhost configuration
- Enhanced SRT server settings
- Additional HTTP hooks for monitoring
- Statistics and monitoring endpoints

### Configuration Sections
- **HLS Input Vhost**: Dedicated vhost for HLS input processing
- **SRT Input Vhost**: Dedicated vhost for SRT input processing
- **Enhanced Default Vhost**: Extended configuration for main streaming
- **Global Monitoring**: System-wide monitoring and statistics

## 6. UI Components

### Enhanced Features Component
The new UI component provides:
- Tabbed interface for all enhanced features
- HLS input management interface
- SRT input management interface
- Bypass transcode task management
- Real-time monitoring dashboard

### Features
- **HLS Input Manager**: Add, configure, and manage HLS input streams
- **SRT Input Manager**: Configure SRT input ports and monitor connections
- **Bypass Transcode Manager**: Create and manage bypass transcoding tasks
- **Monitoring Dashboard**: Real-time metrics and historical charts

## 7. Installation and Setup

### Prerequisites
- Oryx platform running
- Redis server accessible
- SRS server with enhanced configuration

### Configuration Steps
1. **Update SRS Configuration**:
   ```bash
   cp platform/containers/conf/srs.enhanced.conf platform/containers/conf/srs.server.conf
   ```

2. **Restart SRS Server**:
   ```bash
   # SRS will automatically reload the new configuration
   ```

3. **Access Enhanced Features**:
   - Navigate to the Enhanced Features tab in the Oryx UI
   - Configure HLS inputs, SRT inputs, and bypass transcode tasks
   - Monitor system performance through the monitoring dashboard

## 8. Usage Examples

### Adding HLS Input
1. Navigate to "HLS Input" tab
2. Enter stream name and HLS URL
3. Click "Add" to create the input
4. Click "Start" to begin processing

### Adding SRT Input
1. Navigate to "SRT Input" tab
2. Enter port name and port number
3. Click "Add" to create the input
4. Connect SRT clients to the specified port

### Creating Bypass Transcode Task
1. Navigate to "Bypass Transcode" tab
2. Configure input and output parameters
3. Select bypass mode (passthrough or filter)
4. Click "Add Task" to create the task

### Monitoring System
1. Navigate to "Monitoring" tab
2. View real-time metrics
3. Select time period for historical data
4. Analyze bandwidth trends and stream counts

## 9. Troubleshooting

### Common Issues
1. **HLS Input Not Working**:
   - Verify HLS URL is accessible
   - Check SRS logs for errors
   - Ensure HLS input vhost is enabled

2. **SRT Connection Issues**:
   - Verify port is not blocked by firewall
   - Check SRT client configuration
   - Ensure maximum 2 streams per port

3. **Monitoring Data Missing**:
   - Check Redis connection
   - Verify monitoring is enabled
   - Check SRS statistics endpoints

### Log Locations
- SRS logs: Docker container logs
- Application logs: Platform service logs
- Redis logs: Redis server logs

## 10. Performance Considerations

### Resource Usage
- **Memory**: Additional memory for stream processing
- **CPU**: Minimal CPU usage for bypass transcoding
- **Network**: Bandwidth monitoring overhead

### Optimization Tips
- Use appropriate sampling rates for monitoring
- Configure data retention periods based on needs
- Monitor Redis memory usage for large datasets

## 11. Security Considerations

### Authentication
- All API endpoints require valid management token
- Token-based authentication for all operations
- Secure communication over HTTPS

### Access Control
- Management interface access control
- API endpoint security
- Stream access verification

## 12. Future Enhancements

### Planned Features
- Advanced SCTE-35 filtering
- Custom video metadata handling
- Enhanced analytics and reporting
- Multi-tenant support
- Cloud storage integration

### API Extensions
- WebSocket support for real-time updates
- GraphQL API for complex queries
- REST API versioning
- OpenAPI specification

## Support

For technical support and questions about the enhanced features:
- Check the Oryx documentation
- Review SRS configuration examples
- Monitor system logs for errors
- Contact the development team

---

*This document is part of the Enhanced Oryx Platform documentation.* 