//
// Copyright (c) 2022-2024 Winlin
//
// SPDX-License-Identifier: MIT
//
import React, { useState, useEffect } from 'react';
import {
  Container,
  Row,
  Col,
  Card,
  Button,
  Table,
  Form,
  Modal,
  Alert,
  Badge,
  Spinner,
  Tabs,
  Tab,
  InputGroup,
  FormControl,
  Dropdown,
  ButtonGroup
} from 'react-bootstrap';
import {
  Plus,
  Trash2,
  Play,
  Pause,
  Settings,
  Eye,
  EyeOff,
  RefreshCw,
  Wifi,
  WifiOff,
  Monitor,
  Camera,
  Mic,
  Volume2
} from 'lucide-react';
import axios from 'axios';
import { useTranslation } from 'react-i18next';
import '../stream-control.css';

export default function StreamControl() {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState('inputs');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  // Input Streams State
  const [inputStreams, setInputStreams] = useState([]);
  const [showInputModal, setShowInputModal] = useState(false);
  const [newInput, setNewInput] = useState({
    name: '',
    type: 'rtmp',
    url: '',
    port: 1935,
    enabled: true,
    description: ''
  });

  // Output Streams State
  const [outputStreams, setOutputStreams] = useState([]);
  const [showOutputModal, setShowOutputModal] = useState(false);
  const [newOutput, setNewOutput] = useState({
    name: '',
    type: 'rtmp',
    url: '',
    port: 1935,
    enabled: true,
    description: ''
  });

  // All Streams State
  const [allStreams, setAllStreams] = useState([]);

  useEffect(() => {
    fetchAllStreams();
  }, []);

  const fetchAllStreams = async () => {
    setLoading(true);
    try {
      const [inputsRes, outputsRes, streamsRes] = await Promise.all([
        axios.get('/terraform/v1/streams/inputs'),
        axios.get('/terraform/v1/streams/outputs'),
        axios.get('/terraform/v1/streams/all')
      ]);

      setInputStreams(inputsRes.data.data || []);
      setOutputStreams(outputsRes.data.data || []);
      setAllStreams(streamsRes.data.data || []);
    } catch (err) {
      setError('ไม่สามารถโหลดข้อมูล stream ได้');
      console.error('Error fetching streams:', err);
    } finally {
      setLoading(false);
    }
  };

  const createInputStream = async () => {
    try {
      await axios.post('/terraform/v1/streams/inputs', newInput);
      setSuccess('สร้าง input stream สำเร็จ');
      setShowInputModal(false);
      setNewInput({
        name: '',
        type: 'rtmp',
        url: '',
        port: 1935,
        enabled: true,
        description: ''
      });
      fetchAllStreams();
    } catch (err) {
      setError('ไม่สามารถสร้าง input stream ได้');
      console.error('Error creating input stream:', err);
    }
  };

  const createOutputStream = async () => {
    try {
      await axios.post('/terraform/v1/streams/outputs', newOutput);
      setSuccess('สร้าง output stream สำเร็จ');
      setShowOutputModal(false);
      setNewOutput({
        name: '',
        type: 'rtmp',
        url: '',
        port: 1935,
        enabled: true,
        description: ''
      });
      fetchAllStreams();
    } catch (err) {
      setError('ไม่สามารถสร้าง output stream ได้');
      console.error('Error creating output stream:', err);
    }
  };

  const deleteStream = async (id, type) => {
    if (!window.confirm('คุณแน่ใจหรือไม่ที่จะลบ stream นี้?')) return;

    try {
      await axios.delete(`/terraform/v1/streams/${type}/${id}`);
      setSuccess('ลบ stream สำเร็จ');
      fetchAllStreams();
    } catch (err) {
      setError('ไม่สามารถลบ stream ได้');
      console.error('Error deleting stream:', err);
    }
  };

  const toggleStream = async (id, type, enabled) => {
    try {
      await axios.put(`/terraform/v1/streams/${type}/${id}`, { enabled: !enabled });
      setSuccess(`${enabled ? 'ปิด' : 'เปิด'} stream สำเร็จ`);
      fetchAllStreams();
    } catch (err) {
      setError('ไม่สามารถเปลี่ยนสถานะ stream ได้');
      console.error('Error toggling stream:', err);
    }
  };

  const getStreamTypeIcon = (type) => {
    switch (type) {
      case 'rtmp': return <Monitor className="text-primary" />;
      case 'srt': return <Wifi className="text-success" />;
      case 'hls': return <Camera className="text-info" />;
      case 'webrtc': return <Mic className="text-warning" />;
      default: return <Volume2 className="text-secondary" />;
    }
  };

  const getStatusBadge = (status) => {
    switch (status) {
      case 'active': return <Badge bg="success">ใช้งาน</Badge>;
      case 'inactive': return <Badge bg="secondary">ไม่ใช้งาน</Badge>;
      case 'error': return <Badge bg="danger">ข้อผิดพลาด</Badge>;
      case 'connecting': return <Badge bg="warning">กำลังเชื่อมต่อ</Badge>;
      default: return <Badge bg="light" text="dark">ไม่ทราบ</Badge>;
    }
  };

  const InputStreamsTab = () => (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-3">
        <h5>Input Streams</h5>
        <Button className="stream-add-btn" onClick={() => setShowInputModal(true)}>
          <Plus size={16} className="me-2" />
          เพิ่ม Input Stream
        </Button>
      </div>

      <Table responsive striped hover>
        <thead>
          <tr>
            <th>ชื่อ</th>
            <th>ประเภท</th>
            <th>URL/Port</th>
            <th>สถานะ</th>
            <th>คำอธิบาย</th>
            <th>การดำเนินการ</th>
          </tr>
        </thead>
        <tbody>
          {inputStreams.map((stream) => (
            <tr key={stream.id}>
              <td>
                <div className="d-flex align-items-center">
                  <span className="stream-type-icon">{getStreamTypeIcon(stream.type)}</span>
                  <span className="ms-2">{stream.name}</span>
                </div>
              </td>
              <td>
                <Badge bg="info" className="stream-status-badge">{stream.type.toUpperCase()}</Badge>
              </td>
              <td>
                <code className="stream-code">{stream.url || `Port: ${stream.port}`}</code>
              </td>
              <td>{getStatusBadge(stream.status)}</td>
              <td>{stream.description || '-'}</td>
              <td>
                <ButtonGroup size="sm" className="stream-actions">
                  <Button
                    variant={stream.enabled ? "warning" : "success"}
                    onClick={() => toggleStream(stream.id, 'inputs', stream.enabled)}
                    title={stream.enabled ? "ปิด" : "เปิด"}
                  >
                    {stream.enabled ? <Pause size={14} /> : <Play size={14} />}
                  </Button>
                  <Button
                    variant="danger"
                    onClick={() => deleteStream(stream.id, 'inputs')}
                    title="ลบ"
                  >
                    <Trash2 size={14} />
                  </Button>
                </ButtonGroup>
              </td>
            </tr>
          ))}
        </tbody>
      </Table>

      {inputStreams.length === 0 && (
        <div className="stream-empty-state">
          <Monitor size={48} className="mb-2" />
          <p>ยังไม่มี Input Stream</p>
        </div>
      )}
    </div>
  );

  const OutputStreamsTab = () => (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-3">
        <h5>Output Streams</h5>
        <Button className="stream-add-btn" onClick={() => setShowOutputModal(true)}>
          <Plus size={16} className="me-2" />
          เพิ่ม Output Stream
        </Button>
      </div>

      <Table responsive striped hover>
        <thead>
          <tr>
            <th>ชื่อ</th>
            <th>ประเภท</th>
            <th>URL/Port</th>
            <th>สถานะ</th>
            <th>คำอธิบาย</th>
            <th>การดำเนินการ</th>
          </tr>
        </thead>
        <tbody>
          {outputStreams.map((stream) => (
            <tr key={stream.id}>
              <td>
                <div className="d-flex align-items-center">
                  <span className="stream-type-icon">{getStreamTypeIcon(stream.type)}</span>
                  <span className="ms-2">{stream.name}</span>
                </div>
              </td>
              <td>
                <Badge bg="info" className="stream-status-badge">{stream.type.toUpperCase()}</Badge>
              </td>
              <td>
                <code className="stream-code">{stream.url || `Port: ${stream.port}`}</code>
              </td>
              <td>{getStatusBadge(stream.status)}</td>
              <td>{stream.description || '-'}</td>
              <td>
                <ButtonGroup size="sm" className="stream-actions">
                  <Button
                    variant={stream.enabled ? "warning" : "success"}
                    onClick={() => toggleStream(stream.id, 'outputs', stream.enabled)}
                    title={stream.enabled ? "ปิด" : "เปิด"}
                  >
                    {stream.enabled ? <Pause size={14} /> : <Play size={14} />}
                  </Button>
                  <Button
                    variant="danger"
                    onClick={() => deleteStream(stream.id, 'outputs')}
                    title="ลบ"
                  >
                    <Trash2 size={14} />
                  </Button>
                </ButtonGroup>
              </td>
            </tr>
          ))}
        </tbody>
      </Table>

      {outputStreams.length === 0 && (
        <div className="stream-empty-state">
          <Camera size={48} className="mb-2" />
          <p>ยังไม่มี Output Stream</p>
        </div>
      )}
    </div>
  );

  const AllStreamsTab = () => (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-3">
        <h5>All Streams</h5>
        <Button className="stream-refresh-btn" onClick={fetchAllStreams}>
          <RefreshCw size={16} className="me-2" />
          รีเฟรช
        </Button>
      </div>

      <Table responsive striped hover>
        <thead>
          <tr>
            <th>ชื่อ</th>
            <th>ประเภท</th>
            <th>ทิศทาง</th>
            <th>URL/Port</th>
            <th>สถานะ</th>
            <th>การเชื่อมต่อ</th>
            <th>การดำเนินการ</th>
          </tr>
        </thead>
        <tbody>
          {allStreams.map((stream) => (
            <tr key={stream.id}>
              <td>
                <div className="d-flex align-items-center">
                  <span className="stream-type-icon">{getStreamTypeIcon(stream.type)}</span>
                  <span className="ms-2">{stream.name}</span>
                </div>
              </td>
              <td>
                <Badge bg="info" className="stream-status-badge">{stream.type.toUpperCase()}</Badge>
              </td>
              <td>
                <Badge bg={stream.direction === 'input' ? 'primary' : 'secondary'} className="stream-status-badge">
                  {stream.direction === 'input' ? 'Input' : 'Output'}
                </Badge>
              </td>
              <td>
                <code className="stream-code">{stream.url || `Port: ${stream.port}`}</code>
              </td>
              <td>{getStatusBadge(stream.status)}</td>
              <td>
                {stream.connected ? (
                  <Badge bg="success" className="stream-status-badge">
                    <Wifi size={12} className="me-1" />
                    เชื่อมต่อ
                  </Badge>
                ) : (
                  <Badge bg="danger" className="stream-status-badge">
                    <WifiOff size={12} className="me-1" />
                    ไม่เชื่อมต่อ
                  </Badge>
                )}
              </td>
              <td>
                <ButtonGroup size="sm" className="stream-actions">
                  <Button
                    variant="outline-primary"
                    title="ดูรายละเอียด"
                  >
                    <Eye size={14} />
                  </Button>
                  <Button
                    variant={stream.enabled ? "warning" : "success"}
                    onClick={() => toggleStream(stream.id, stream.direction + 's', stream.enabled)}
                    title={stream.enabled ? "ปิด" : "เปิด"}
                  >
                    {stream.enabled ? <Pause size={14} /> : <Play size={14} />}
                  </Button>
                </ButtonGroup>
              </td>
            </tr>
          ))}
        </tbody>
      </Table>

      {allStreams.length === 0 && (
        <div className="stream-empty-state">
          <Volume2 size={48} className="mb-2" />
          <p>ยังไม่มี Stream</p>
        </div>
      )}
    </div>
  );

  return (
    <Container fluid className="stream-control-container">
      <Row>
        <Col>
          <Card className="stream-table">
            <Card.Header className="stream-control-header">
              <div className="d-flex justify-content-between align-items-center">
                <h4 className="mb-0">
                  <Settings className="me-2" />
                  Stream Control
                </h4>
                <Button className="stream-refresh-btn" onClick={fetchAllStreams} disabled={loading}>
                  {loading ? <Spinner size="sm" /> : <RefreshCw size={16} />}
                </Button>
              </div>
            </Card.Header>
            <Card.Body>
              {error && (
                <Alert variant="danger" dismissible onClose={() => setError('')} className="stream-alert">
                  {error}
                </Alert>
              )}
              {success && (
                <Alert variant="success" dismissible onClose={() => setSuccess('')} className="stream-alert">
                  {success}
                </Alert>
              )}

              <Tabs
                activeKey={activeTab}
                onSelect={(k) => setActiveTab(k)}
                className="stream-tabs"
              >
                <Tab eventKey="inputs" title="Input Streams">
                  <InputStreamsTab />
                </Tab>
                <Tab eventKey="outputs" title="Output Streams">
                  <OutputStreamsTab />
                </Tab>
                <Tab eventKey="all" title="All Streams">
                  <AllStreamsTab />
                </Tab>
              </Tabs>
            </Card.Body>
          </Card>
        </Col>
      </Row>

      {/* Input Stream Modal */}
      <Modal show={showInputModal} onHide={() => setShowInputModal(false)} size="lg" className="stream-modal">
        <Modal.Header closeButton>
          <Modal.Title>เพิ่ม Input Stream</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Form>
            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>ชื่อ Stream</Form.Label>
                  <Form.Control
                    type="text"
                    value={newInput.name}
                    onChange={(e) => setNewInput({ ...newInput, name: e.target.value })}
                    placeholder="ชื่อ stream"
                  />
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>ประเภท</Form.Label>
                  <Form.Select
                    value={newInput.type}
                    onChange={(e) => setNewInput({ ...newInput, type: e.target.value })}
                  >
                    <option value="rtmp">RTMP</option>
                    <option value="srt">SRT</option>
                    <option value="hls">HLS</option>
                    <option value="webrtc">WebRTC</option>
                  </Form.Select>
                </Form.Group>
              </Col>
            </Row>
            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>URL</Form.Label>
                  <Form.Control
                    type="text"
                    value={newInput.url}
                    onChange={(e) => setNewInput({ ...newInput, url: e.target.value })}
                    placeholder="rtmp://example.com/live/stream"
                  />
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Port</Form.Label>
                  <Form.Control
                    type="number"
                    value={newInput.port}
                    onChange={(e) => setNewInput({ ...newInput, port: parseInt(e.target.value) })}
                    placeholder="1935"
                  />
                </Form.Group>
              </Col>
            </Row>
            <Form.Group className="mb-3">
              <Form.Label>คำอธิบาย</Form.Label>
              <Form.Control
                as="textarea"
                rows={3}
                value={newInput.description}
                onChange={(e) => setNewInput({ ...newInput, description: e.target.value })}
                placeholder="คำอธิบาย stream"
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Check
                type="checkbox"
                label="เปิดใช้งาน"
                checked={newInput.enabled}
                onChange={(e) => setNewInput({ ...newInput, enabled: e.target.checked })}
              />
            </Form.Group>
          </Form>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowInputModal(false)}>
            ยกเลิก
          </Button>
          <Button variant="primary" onClick={createInputStream}>
            สร้าง Stream
          </Button>
        </Modal.Footer>
      </Modal>

      {/* Output Stream Modal */}
      <Modal show={showOutputModal} onHide={() => setShowOutputModal(false)} size="lg" className="stream-modal">
        <Modal.Header closeButton>
          <Modal.Title>เพิ่ม Output Stream</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Form>
            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>ชื่อ Stream</Form.Label>
                  <Form.Control
                    type="text"
                    value={newOutput.name}
                    onChange={(e) => setNewOutput({ ...newOutput, name: e.target.value })}
                    placeholder="ชื่อ stream"
                  />
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>ประเภท</Form.Label>
                  <Form.Select
                    value={newOutput.type}
                    onChange={(e) => setNewOutput({ ...newOutput, type: e.target.value })}
                  >
                    <option value="rtmp">RTMP</option>
                    <option value="srt">SRT</option>
                    <option value="hls">HLS</option>
                    <option value="webrtc">WebRTC</option>
                  </Form.Select>
                </Form.Group>
              </Col>
            </Row>
            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>URL</Form.Label>
                  <Form.Control
                    type="text"
                    value={newOutput.url}
                    onChange={(e) => setNewOutput({ ...newOutput, url: e.target.value })}
                    placeholder="rtmp://example.com/live/stream"
                  />
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Port</Form.Label>
                  <Form.Control
                    type="number"
                    value={newOutput.port}
                    onChange={(e) => setNewOutput({ ...newOutput, port: parseInt(e.target.value) })}
                    placeholder="1935"
                  />
                </Form.Group>
              </Col>
            </Row>
            <Form.Group className="mb-3">
              <Form.Label>คำอธิบาย</Form.Label>
              <Form.Control
                as="textarea"
                rows={3}
                value={newOutput.description}
                onChange={(e) => setNewOutput({ ...newOutput, description: e.target.value })}
                placeholder="คำอธิบาย stream"
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Check
                type="checkbox"
                label="เปิดใช้งาน"
                checked={newOutput.enabled}
                onChange={(e) => setNewOutput({ ...newOutput, enabled: e.target.checked })}
              />
            </Form.Group>
          </Form>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowOutputModal(false)}>
            ยกเลิก
          </Button>
          <Button variant="primary" onClick={createOutputStream}>
            สร้าง Stream
          </Button>
        </Modal.Footer>
      </Modal>
    </Container>
  );
}
