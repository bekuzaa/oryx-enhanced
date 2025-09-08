# Stream Control - ระบบควบคุมสตรีม

## ภาพรวม
หน้า Stream Control เป็นระบบจัดการสตรีมแบบครบวงจรที่ช่วยให้ผู้ใช้สามารถจัดการ Input และ Output streams ได้อย่างง่ายดายผ่าน Web Interface

## ฟีเจอร์หลัก

### 1. การจัดการ Input Streams
- **เพิ่ม Input Stream**: สร้างสตรีมใหม่สำหรับรับข้อมูล
- **ลบ Input Stream**: ลบสตรีมที่ไม่ต้องการ
- **เปิด/ปิด Stream**: ควบคุมการทำงานของสตรีม
- **แสดงสถานะ**: ดูสถานะการเชื่อมต่อแบบ Real-time

### 2. การจัดการ Output Streams
- **เพิ่ม Output Stream**: สร้างสตรีมใหม่สำหรับส่งข้อมูล
- **ลบ Output Stream**: ลบสตรีมที่ไม่ต้องการ
- **เปิด/ปิด Stream**: ควบคุมการทำงานของสตรีม
- **แสดงสถานะ**: ดูสถานะการเชื่อมต่อแบบ Real-time

### 3. การดู All Streams
- **รวมทุกสตรีม**: ดู Input และ Output streams ในหน้าเดียว
- **สถานะการเชื่อมต่อ**: แสดงสถานะการเชื่อมต่อของแต่ละสตรีม
- **รีเฟรชข้อมูล**: อัปเดตข้อมูลแบบ Real-time

## ประเภทสตรีมที่รองรับ

### 1. RTMP
- **Port**: 1935 (default)
- **ใช้สำหรับ**: การสตรีมแบบ Real-time
- **ตัวอย่าง URL**: `rtmp://example.com/live/stream`

### 2. SRT
- **Port**: 10080 (with StreamID), 10081/10082 (without StreamID)
- **ใช้สำหรับ**: การสตรีมคุณภาพสูงแบบ Low-latency
- **ตัวอย่าง URL**: `srt://example.com:10080?streamid=stream1`

### 3. HLS
- **Port**: 8080 (default)
- **ใช้สำหรับ**: การสตรีมแบบ HTTP Live Streaming
- **ตัวอย่าง URL**: `http://example.com:8080/hls/stream.m3u8`

### 4. WebRTC
- **Port**: 8000 (default)
- **ใช้สำหรับ**: การสตรีมแบบ Real-time Communication
- **ตัวอย่าง URL**: `webrtc://example.com:8000/stream`

## API Endpoints

### Input Streams
```
GET    /terraform/v1/streams/inputs     - ดึงรายการ Input streams
POST   /terraform/v1/streams/inputs     - สร้าง Input stream ใหม่
PUT    /terraform/v1/streams/inputs/{id} - อัปเดต Input stream
DELETE /terraform/v1/streams/inputs/{id} - ลบ Input stream
```

### Output Streams
```
GET    /terraform/v1/streams/outputs     - ดึงรายการ Output streams
POST   /terraform/v1/streams/outputs     - สร้าง Output stream ใหม่
PUT    /terraform/v1/streams/outputs/{id} - อัปเดต Output stream
DELETE /terraform/v1/streams/outputs/{id} - ลบ Output stream
```

### All Streams
```
GET    /terraform/v1/streams/all        - ดึงรายการสตรีมทั้งหมด
```

## วิธีการใช้งาน

### 1. การเพิ่ม Input Stream
1. คลิกที่แท็บ "Input Streams"
2. คลิกปุ่ม "เพิ่ม Input Stream"
3. กรอกข้อมูล:
   - **ชื่อ Stream**: ชื่อที่ต้องการ
   - **ประเภท**: เลือกประเภทสตรีม (RTMP, SRT, HLS, WebRTC)
   - **URL**: URL ของสตรีม
   - **Port**: พอร์ตที่ใช้ (จะตั้งค่า default อัตโนมัติ)
   - **คำอธิบาย**: คำอธิบายเพิ่มเติม
4. คลิก "สร้าง Stream"

### 2. การเพิ่ม Output Stream
1. คลิกที่แท็บ "Output Streams"
2. คลิกปุ่ม "เพิ่ม Output Stream"
3. กรอกข้อมูลเหมือนกับ Input Stream
4. คลิก "สร้าง Stream"

### 3. การจัดการสตรีม
- **เปิด/ปิด**: คลิกปุ่ม Play/Pause
- **ลบ**: คลิกปุ่ม Trash
- **ดูรายละเอียด**: คลิกปุ่ม Eye (ใน All Streams)

## สถานะสตรีม

### สถานะการทำงาน
- **Active**: สตรีมกำลังทำงาน
- **Inactive**: สตรีมไม่ทำงาน
- **Error**: เกิดข้อผิดพลาด
- **Connecting**: กำลังเชื่อมต่อ

### สถานะการเชื่อมต่อ
- **เชื่อมต่อ**: สตรีมเชื่อมต่อสำเร็จ
- **ไม่เชื่อมต่อ**: สตรีมไม่สามารถเชื่อมต่อได้

## การตั้งค่า SRT Ports

### Port 10080 - SRT with StreamID
- ใช้สำหรับสตรีมที่ต้องการ StreamID
- ตัวอย่าง: `srt://server:10080?streamid=#!::r=live/stream1`

### Port 10081 - SRT without StreamID (Stream 1)
- ใช้สำหรับสตรีมที่ไม่ต้องการ StreamID
- ตัวอย่าง: `srt://server:10081`

### Port 10082 - SRT without StreamID (Stream 2)
- ใช้สำหรับสตรีมที่ไม่ต้องการ StreamID
- ตัวอย่าง: `srt://server:10082`

## การแก้ไขปัญหา

### สตรีมไม่สามารถเชื่อมต่อได้
1. ตรวจสอบ URL และ Port
2. ตรวจสอบ Firewall settings
3. ตรวจสอบสถานะของ SRS server

### สตรีมแสดงสถานะ Error
1. ตรวจสอบการตั้งค่า
2. ดู Log files
3. ตรวจสอบ Network connectivity

## การพัฒนา

### ไฟล์ที่เกี่ยวข้อง
- **Frontend**: `ui/src/pages/StreamControl.js`
- **Backend**: `platform/stream-control.go`
- **CSS**: `ui/src/stream-control.css`
- **Routing**: `ui/src/App.js`

### การเพิ่มฟีเจอร์ใหม่
1. อัปเดต Backend API ใน `stream-control.go`
2. อัปเดต Frontend UI ใน `StreamControl.js`
3. เพิ่ม CSS styles ใน `stream-control.css`
4. ทดสอบการทำงาน

## การทดสอบ

### ทดสอบ API
```bash
# ดึงรายการ Input streams
curl -X GET http://localhost:2022/terraform/v1/streams/inputs \
  -H "Content-Type: application/json" \
  -d '{"token": "your_token"}'

# สร้าง Input stream ใหม่
curl -X POST http://localhost:2022/terraform/v1/streams/inputs \
  -H "Content-Type: application/json" \
  -d '{
    "token": "your_token",
    "name": "Test Stream",
    "type": "rtmp",
    "url": "rtmp://example.com/live/test",
    "port": 1935,
    "enabled": true,
    "description": "Test stream"
  }'
```

### ทดสอบ UI
1. เข้าสู่ระบบ Oryx
2. ไปที่หน้า "ควบคุมสตรีม"
3. ทดสอบการเพิ่ม/ลบ/แก้ไขสตรีม
4. ตรวจสอบการแสดงผลในแต่ละแท็บ

## การบำรุงรักษา

### การสำรองข้อมูล
- ข้อมูลสตรีมจะถูกเก็บใน Redis
- ควรสำรองข้อมูล Redis เป็นประจำ

### การอัปเดต
- อัปเดต Frontend และ Backend พร้อมกัน
- ทดสอบการทำงานหลังการอัปเดต
- ตรวจสอบ API compatibility

## การสนับสนุน

หากพบปัญหาหรือต้องการความช่วยเหลือ:
1. ตรวจสอบ Log files
2. ดู Documentation
3. ติดต่อทีมพัฒนา
