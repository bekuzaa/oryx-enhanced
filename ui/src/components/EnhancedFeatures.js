import React, { useState, useEffect } from 'react';
import {
  Box,
  Tabs,
  Tab,
  Typography,
  Card,
  CardContent,
  Grid,
  Button,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Switch,
  FormControlLabel,
  Alert,
  CircularProgress,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  IconButton,
  Tooltip,
} from '@mui/material';
import {
  PlayArrow,
  Stop,
  Delete,
  Edit,
  Refresh,
  Timeline,
  ShowChart,
  Settings,
} from '@mui/icons-material';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip as RechartsTooltip, ResponsiveContainer } from 'recharts';

// HLS Input Management Component
const HLSInputManager = () => {
  const [inputs, setInputs] = useState([]);
  const [loading, setLoading] = useState(false);
  const [newInput, setNewInput] = useState({ name: '', url: '', enabled: true });
  const [error, setError] = useState('');

  const fetchInputs = async () => {
    setLoading(true);
    try {
      const response = await fetch('/terraform/v1/hls/input/query', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token: localStorage.getItem('mgmt_token') }),
      });
      const data = await response.json();
      if (data.code === 0) {
        setInputs(data.data || []);
      }
    } catch (err) {
      setError('Failed to fetch HLS inputs');
    }
    setLoading(false);
  };

  const createInput = async () => {
    try {
      const response = await fetch('/terraform/v1/hls/input/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          token: localStorage.getItem('mgmt_token'),
          ...newInput,
        }),
      });
      const data = await response.json();
      if (data.code === 0) {
        setNewInput({ name: '', url: '', enabled: true });
        fetchInputs();
      }
    } catch (err) {
      setError('Failed to create HLS input');
    }
  };

  const deleteInput = async (id) => {
    try {
      await fetch('/terraform/v1/hls/input/delete', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          token: localStorage.getItem('mgmt_token'),
          id,
        }),
      });
      fetchInputs();
    } catch (err) {
      setError('Failed to delete HLS input');
    }
  };

  useEffect(() => {
    fetchInputs();
  }, []);

  return (
    <Card>
      <CardContent>
        <Typography variant="h6" gutterBottom>
          HLS Input Management
        </Typography>
        
        <Box sx={{ mb: 3 }}>
          <Grid container spacing={2} alignItems="center">
            <Grid item xs={3}>
              <TextField
                fullWidth
                label="Name"
                value={newInput.name}
                onChange={(e) => setNewInput({ ...newInput, name: e.target.value })}
              />
            </Grid>
            <Grid item xs={6}>
              <TextField
                fullWidth
                label="HLS URL"
                value={newInput.url}
                onChange={(e) => setNewInput({ ...newInput, url: e.target.value })}
              />
            </Grid>
            <Grid item xs={2}>
              <FormControlLabel
                control={
                  <Switch
                    checked={newInput.enabled}
                    onChange={(e) => setNewInput({ ...newInput, enabled: e.target.checked })}
                  />
                }
                label="Enabled"
              />
            </Grid>
            <Grid item xs={1}>
              <Button variant="contained" onClick={createInput}>
                Add
              </Button>
            </Grid>
          </Grid>
        </Box>

        {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>URL</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Stream Count</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {inputs.map((input) => (
                <TableRow key={input.id}>
                  <TableCell>{input.name}</TableCell>
                  <TableCell>{input.url}</TableCell>
                  <TableCell>
                    <Chip
                      label={input.status}
                      color={input.status === 'active' ? 'success' : 'default'}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>{input.streamCount}</TableCell>
                  <TableCell>
                    <Tooltip title="Start">
                      <IconButton size="small" color="primary">
                        <PlayArrow />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="Stop">
                      <IconButton size="small" color="secondary">
                        <Stop />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="Delete">
                      <IconButton size="small" color="error" onClick={() => deleteInput(input.id)}>
                        <Delete />
                      </IconButton>
                    </Tooltip>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </CardContent>
    </Card>
  );
};

// SRT Input Management Component
const SRTInputManager = () => {
  const [inputs, setInputs] = useState([]);
  const [streams, setStreams] = useState([]);
  const [loading, setLoading] = useState(false);
  const [newInput, setNewInput] = useState({ 
    name: '', 
    port: 10080, 
    portNoStreamId1: 10081, 
    portNoStreamId2: 10082, 
    enabled: true 
  });
  const [error, setError] = useState('');

  const fetchInputs = async () => {
    setLoading(true);
    try {
      const response = await fetch('/terraform/v1/srt/input/query', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token: localStorage.getItem('mgmt_token') }),
      });
      const data = await response.json();
      if (data.code === 0) {
        setInputs(data.data || []);
      }
    } catch (err) {
      setError('Failed to fetch SRT inputs');
    }
    setLoading(false);
  };

  const createInput = async () => {
    try {
      const response = await fetch('/terraform/v1/srt/input/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          token: localStorage.getItem('mgmt_token'),
          ...newInput,
        }),
      });
      const data = await response.json();
      if (data.code === 0) {
        setNewInput({ 
          name: '', 
          port: 10080, 
          portNoStreamId1: 10081, 
          portNoStreamId2: 10082, 
          enabled: true 
        });
        fetchInputs();
      }
    } catch (err) {
      setError('Failed to create SRT input');
    }
  };

  const deleteInput = async (id) => {
    try {
      await fetch('/terraform/v1/srt/input/delete', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          token: localStorage.getItem('mgmt_token'),
          id,
        }),
      });
      fetchInputs();
    } catch (err) {
      setError('Failed to delete SRT input');
    }
  };

  useEffect(() => {
    fetchInputs();
  }, []);

  return (
    <Card>
      <CardContent>
        <Typography variant="h6" gutterBottom>
          SRT Input Management (Max 2 Streams)
        </Typography>
        
        <Box sx={{ mb: 3 }}>
          <Grid container spacing={2} alignItems="center">
            <Grid item xs={4}>
              <TextField
                fullWidth
                label="Name"
                value={newInput.name}
                onChange={(e) => setNewInput({ ...newInput, name: e.target.value })}
              />
            </Grid>
            <Grid item xs={2}>
              <TextField
                fullWidth
                label="Port (StreamID)"
                type="number"
                value={newInput.port}
                onChange={(e) => setNewInput({ ...newInput, port: parseInt(e.target.value) })}
                helperText="Default: 10080"
              />
            </Grid>
            <Grid item xs={2}>
              <TextField
                fullWidth
                label="Port (No StreamID 1)"
                type="number"
                value={newInput.portNoStreamId1}
                onChange={(e) => setNewInput({ ...newInput, portNoStreamId1: parseInt(e.target.value) })}
                helperText="Default: 10081"
              />
            </Grid>
            <Grid item xs={2}>
              <TextField
                fullWidth
                label="Port (No StreamID 2)"
                type="number"
                value={newInput.portNoStreamId2}
                onChange={(e) => setNewInput({ ...newInput, portNoStreamId2: parseInt(e.target.value) })}
                helperText="Default: 10082"
              />
            </Grid>
            <Grid item xs={3}>
              <FormControlLabel
                control={
                  <Switch
                    checked={newInput.enabled}
                    onChange={(e) => setNewInput({ ...newInput, enabled: e.target.checked })}
                  />
                }
                label="Enabled"
              />
            </Grid>
            <Grid item xs={2}>
              <Button variant="contained" onClick={createInput}>
                Add
              </Button>
            </Grid>
          </Grid>
        </Box>

        {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Port (StreamID)</TableCell>
                <TableCell>Port (No StreamID 1)</TableCell>
                <TableCell>Port (No StreamID 2)</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Stream Count</TableCell>
                <TableCell>Max Streams</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {inputs.map((input) => (
                <TableRow key={input.id}>
                  <TableCell>{input.name}</TableCell>
                  <TableCell>{input.port}</TableCell>
                  <TableCell>{input.portNoStreamId1 || 'N/A'}</TableCell>
                  <TableCell>{input.portNoStreamId2 || 'N/A'}</TableCell>
                  <TableCell>
                    <Chip
                      label={input.status}
                      color={input.status === 'active' ? 'success' : 'default'}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>{input.streamCount}</TableCell>
                  <TableCell>{input.maxStreams}</TableCell>
                  <TableCell>
                    <Tooltip title="Start">
                      <IconButton size="small" color="primary">
                        <PlayArrow />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="Stop">
                      <IconButton size="small" color="secondary">
                        <Stop />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="Delete">
                      <IconButton size="small" color="error" onClick={() => deleteInput(input.id)}>
                        <Delete />
                      </IconButton>
                    </Tooltip>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </CardContent>
    </Card>
  );
};

// Bypass Transcode Management Component
const BypassTranscodeManager = () => {
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(false);
  const [newTask, setNewTask] = useState({
    name: '',
    inputType: 'hls',
    inputUrl: '',
    outputType: 'rtmp',
    outputUrl: '',
    bypassMode: 'passthrough',
    enabled: true,
  });
  const [error, setError] = useState('');

  const fetchTasks = async () => {
    setLoading(true);
    try {
      const response = await fetch('/terraform/v1/bypass/transcode/query', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token: localStorage.getItem('mgmt_token') }),
      });
      const data = await response.json();
      if (data.code === 0) {
        setTasks(data.data || []);
      }
    } catch (err) {
      setError('Failed to fetch bypass transcode tasks');
    }
    setLoading(false);
  };

  const createTask = async () => {
    try {
      const response = await fetch('/terraform/v1/bypass/transcode/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          token: localStorage.getItem('mgmt_token'),
          ...newTask,
        }),
      });
      const data = await response.json();
      if (data.code === 0) {
        setNewTask({
          name: '',
          inputType: 'hls',
          inputUrl: '',
          outputType: 'rtmp',
          outputUrl: '',
          bypassMode: 'passthrough',
          enabled: true,
        });
        fetchTasks();
      }
    } catch (err) {
      setError('Failed to create bypass transcode task');
    }
  };

  const deleteTask = async (id) => {
    try {
      await fetch('/terraform/v1/bypass/transcode/delete', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          token: localStorage.getItem('mgmt_token'),
          id,
        }),
      });
      fetchTasks();
    } catch (err) {
      setError('Failed to delete bypass transcode task');
    }
  };

  useEffect(() => {
    fetchTasks();
  }, []);

  return (
    <Card>
      <CardContent>
        <Typography variant="h6" gutterBottom>
          Bypass Transcode Management (No FFmpeg Re-encoding)
        </Typography>
        
        <Box sx={{ mb: 3 }}>
          <Grid container spacing={2}>
            <Grid item xs={3}>
              <TextField
                fullWidth
                label="Task Name"
                value={newTask.name}
                onChange={(e) => setNewTask({ ...newTask, name: e.target.value })}
              />
            </Grid>
            <Grid item xs={2}>
              <FormControl fullWidth>
                <InputLabel>Input Type</InputLabel>
                <Select
                  value={newTask.inputType}
                  label="Input Type"
                  onChange={(e) => setNewTask({ ...newTask, inputType: e.target.value })}
                >
                  <MenuItem value="hls">HLS</MenuItem>
                  <MenuItem value="srt">SRT</MenuItem>
                  <MenuItem value="rtmp">RTMP</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={3}>
              <TextField
                fullWidth
                label="Input URL"
                value={newTask.inputUrl}
                onChange={(e) => setNewTask({ ...newTask, inputUrl: e.target.value })}
              />
            </Grid>
            <Grid item xs={2}>
              <FormControl fullWidth>
                <InputLabel>Output Type</InputLabel>
                <Select
                  value={newTask.outputType}
                  label="Output Type"
                  onChange={(e) => setNewTask({ ...newTask, outputType: e.target.value })}
                >
                  <MenuItem value="rtmp">RTMP</MenuItem>
                  <MenuItem value="hls">HLS</MenuItem>
                  <MenuItem value="srt">SRT</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={2}>
              <Button variant="contained" onClick={createTask}>
                Add Task
              </Button>
            </Grid>
          </Grid>
          <Grid container spacing={2} sx={{ mt: 1 }}>
            <Grid item xs={3}>
              <TextField
                fullWidth
                label="Output URL"
                value={newTask.outputUrl}
                onChange={(e) => setNewTask({ ...newTask, outputUrl: e.target.value })}
              />
            </Grid>
            <Grid item xs={3}>
              <FormControl fullWidth>
                <InputLabel>Bypass Mode</InputLabel>
                <Select
                  value={newTask.bypassMode}
                  label="Bypass Mode"
                  onChange={(e) => setNewTask({ ...newTask, bypassMode: e.target.value })}
                >
                  <MenuItem value="passthrough">Passthrough</MenuItem>
                  <MenuItem value="filter">Filter</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={3}>
              <FormControlLabel
                control={
                  <Switch
                    checked={newTask.enabled}
                    onChange={(e) => setNewTask({ ...newTask, enabled: e.target.checked })}
                  />
                }
                label="Enabled"
              />
            </Grid>
          </Grid>
        </Box>

        {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Input</TableCell>
                <TableCell>Output</TableCell>
                <TableCell>Mode</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {tasks.map((task) => (
                <TableRow key={task.id}>
                  <TableCell>{task.name}</TableCell>
                  <TableCell>{task.inputType}: {task.inputUrl}</TableCell>
                  <TableCell>{task.outputType}: {task.outputUrl}</TableCell>
                  <TableCell>{task.bypassMode}</TableCell>
                  <TableCell>
                    <Chip
                      label={task.status}
                      color={task.status === 'active' ? 'success' : 'default'}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>
                    <Tooltip title="Start">
                      <IconButton size="small" color="primary">
                        <PlayArrow />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="Stop">
                      <IconButton size="small" color="secondary">
                        <Stop />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="Delete">
                      <IconButton size="small" color="error" onClick={() => deleteTask(task.id)}>
                        <Delete />
                      </IconButton>
                    </Tooltip>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </CardContent>
    </Card>
  );
};

// Monitoring Dashboard Component
const MonitoringDashboard = () => {
  const [metrics, setMetrics] = useState({});
  const [historicalData, setHistoricalData] = useState([]);
  const [loading, setLoading] = useState(false);
  const [period, setPeriod] = useState('daily');
  const [error, setError] = useState('');

  const fetchRealTimeMetrics = async () => {
    try {
      const response = await fetch('/terraform/v1/monitoring/realtime', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token: localStorage.getItem('mgmt_token') }),
      });
      const data = await response.json();
      if (data.code === 0) {
        setMetrics(data.data || {});
      }
    } catch (err) {
      setError('Failed to fetch real-time metrics');
    }
  };

  const fetchHistoricalData = async () => {
    setLoading(true);
    try {
      const response = await fetch('/terraform/v1/monitoring/query', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          token: localStorage.getItem('mgmt_token'),
          type: 'bandwidth',
          period: period,
          startTime: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000), // Last 7 days
          endTime: new Date(),
        }),
      });
      const data = await response.json();
      if (data.code === 0) {
        setHistoricalData(data.data || []);
      }
    } catch (err) {
      setError('Failed to fetch historical data');
    }
    setLoading(false);
  };

  useEffect(() => {
    fetchRealTimeMetrics();
    const interval = setInterval(fetchRealTimeMetrics, 5000); // Update every 5 seconds
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    fetchHistoricalData();
  }, [period]);

  return (
    <Card>
      <CardContent>
        <Typography variant="h6" gutterBottom>
          Monitoring Dashboard
        </Typography>

        <Box sx={{ mb: 3 }}>
          <Grid container spacing={2} alignItems="center">
            <Grid item xs={3}>
              <FormControl fullWidth>
                <InputLabel>Period</InputLabel>
                <Select
                  value={period}
                  label="Period"
                  onChange={(e) => setPeriod(e.target.value)}
                >
                  <MenuItem value="daily">Daily</MenuItem>
                  <MenuItem value="weekly">Weekly</MenuItem>
                  <MenuItem value="monthly">Monthly</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={2}>
              <Button
                variant="outlined"
                startIcon={<Refresh />}
                onClick={fetchHistoricalData}
                disabled={loading}
              >
                Refresh
              </Button>
            </Grid>
          </Grid>
        </Box>

        {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

        {/* Real-time Metrics */}
        <Box sx={{ mb: 3 }}>
          <Typography variant="h6" gutterBottom>
            Real-time Metrics
          </Typography>
          <Grid container spacing={2}>
            <Grid item xs={6}>
              <Card variant="outlined">
                <CardContent>
                  <Typography color="textSecondary" gutterBottom>
                    Current Bandwidth
                  </Typography>
                  <Typography variant="h4">
                    {metrics.bandwidth?.value || 0} {metrics.bandwidth?.unit || 'Mbps'}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
            <Grid item xs={6}>
              <Card variant="outlined">
                <CardContent>
                  <Typography color="textSecondary" gutterBottom>
                    Concurrent Streams
                  </Typography>
                  <Typography variant="h4">
                    {metrics.concurrent_streams?.value || 0}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
          </Grid>
        </Box>

        {/* Historical Chart */}
        <Box sx={{ mb: 3 }}>
          <Typography variant="h6" gutterBottom>
            Bandwidth Trend ({period})
          </Typography>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={historicalData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis
                dataKey="timestamp"
                tickFormatter={(value) => new Date(value).toLocaleDateString()}
              />
              <YAxis />
              <RechartsTooltip
                labelFormatter={(value) => new Date(value).toLocaleString()}
                formatter={(value, name) => [value, name]}
              />
              <Line type="monotone" dataKey="value" stroke="#8884d8" />
            </LineChart>
          </ResponsiveContainer>
        </Box>
      </CardContent>
    </Card>
  );
};

// Main Enhanced Features Component
const EnhancedFeatures = () => {
  const [activeTab, setActiveTab] = useState(0);

  const handleTabChange = (event, newValue) => {
    setActiveTab(newValue);
  };

  return (
    <Box sx={{ width: '100%' }}>
      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 3 }}>
        <Tabs value={activeTab} onChange={handleTabChange}>
          <Tab label="HLS Input" icon={<PlayArrow />} />
          <Tab label="SRT Input" icon={<Settings />} />
          <Tab label="Bypass Transcode" icon={<Timeline />} />
          <Tab label="Monitoring" icon={<ShowChart />} />
        </Tabs>
      </Box>

      {activeTab === 0 && <HLSInputManager />}
      {activeTab === 1 && <SRTInputManager />}
      {activeTab === 2 && <BypassTranscodeManager />}
      {activeTab === 3 && <MonitoringDashboard />}
    </Box>
  );
};

export default EnhancedFeatures; 