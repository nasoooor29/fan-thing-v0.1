// API Configuration
const API_URL = '';

// State
let points = [];

let chart = null;

// Initialize the application
async function init() {
    await loadConfig(); // Auto-load on startup
    renderPointsList();
    updateChart();
    setupEventListeners();
}

// Setup event listeners
function setupEventListeners() {
    document.getElementById('add-point-btn').addEventListener('click', addPoint);
    document.getElementById('update-curve-btn').addEventListener('click', updateChart);
    document.getElementById('interpolation-mode').addEventListener('change', updateChart);
}

// Render the points list
function renderPointsList() {
    const pointsList = document.getElementById('points-list');
    pointsList.innerHTML = '';

    points.forEach((point, index) => {
        const pointDiv = document.createElement('div');
        pointDiv.className = 'point-item';
        pointDiv.innerHTML = `
            <div class="points-details">
                <div class="point-inputs">
                    <div class="input-group">
                        <label>Temperature (°C):</label>
                        <input 
                            type="number" 
                            min="0" 
                            max="100" 
                            value="${point.temperature}"
                            data-index="${index}"
                            data-field="temperature"
                            class="point-input"
                        >
                    </div>
                    <div class="input-group">
                        <label>Fan Speed (%):</label>
                        <input 
                            type="number" 
                            min="0" 
                            max="100" 
                            value="${point.fanSpeed}"
                            data-index="${index}"
                            data-field="fanSpeed"
                            class="point-input"
                        >
                    </div>
                </div>
                <div class="point-actions">
                    <button class="btn btn-danger" data-index="${index}">Remove</button>
                </div>
            </div>
        `;
        pointsList.appendChild(pointDiv);
    });

    // Add event listeners for inputs
    document.querySelectorAll('.point-input').forEach(input => {
        input.addEventListener('change', handlePointUpdate);
    });

    // Add event listeners for remove buttons
    document.querySelectorAll('.btn-danger').forEach(button => {
        button.addEventListener('click', handlePointRemove);
    });
}

// Handle point update
function handlePointUpdate(event) {
    const index = parseInt(event.target.dataset.index);
    const field = event.target.dataset.field;
    const value = parseFloat(event.target.value);

    if (!isNaN(value)) {
        points[index][field] = value;
    }
}

// Handle point removal
function handlePointRemove(event) {
    const index = parseInt(event.target.dataset.index);
    points.splice(index, 1);
    renderPointsList();
    updateChart();
}

// Add a new point
function addPoint() {
    points.push({ temperature: 50, fanSpeed: 50 });
    renderPointsList();
}

// Update the chart by fetching data from backend (also auto-saves)
async function updateChart() {
    const interpolationMode = document.getElementById('interpolation-mode').value;

    try {
        const response = await fetch(`/api/generate-curve`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                points: points,
                interpolationMode: interpolationMode
            })
        });

        if (!response.ok) {
            throw new Error('Failed to generate curve data');
        }

        const data = await response.json();
        renderChart(data);
    } catch (error) {
        console.error('Error updating chart:', error);
        alert('Failed to update chart. Make sure the backend server is running.');
    }
}

// Render the chart
function renderChart(data) {
    const ctx = document.getElementById('fan-chart').getContext('2d');

    // Extract data for chart
    const curveLabels = data.curveData.map(point => point.x);
    const curveValues = data.curveData.map(point => point.y);
    const controlPointsX = data.controlPoints.map(point => point.temperature);
    const controlPointsY = data.controlPoints.map(point => point.fanSpeed);

    // Destroy existing chart if it exists
    if (chart) {
        chart.destroy();
    }

    // Create new chart
    chart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: curveLabels,
            datasets: [
                {
                    label: 'Fan Speed',
                    data: curveValues,
                    borderColor: 'rgb(59, 130, 246)',
                    backgroundColor: 'rgba(59, 130, 246, 0.1)',
                    fill: true,
                    tension: 0,
                    pointRadius: 0,
                    borderWidth: 2
                },
                {
                    label: 'Control Points',
                    data: controlPointsX.map((x, i) => ({ x: x, y: controlPointsY[i] })),
                    borderColor: 'rgb(239, 68, 68)',
                    backgroundColor: 'rgb(239, 68, 68)',
                    pointRadius: 6,
                    pointHoverRadius: 8,
                    showLine: false,
                    type: 'scatter'
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                x: {
                    title: {
                        display: true,
                        text: 'Temperature (°C)'
                    },
                    min: 0,
                    max: 100
                },
                y: {
                    title: {
                        display: true,
                        text: 'Fan Speed (%)'
                    },
                    min: 0,
                    max: 100
                }
            },
            plugins: {
                legend: {
                    display: true,
                    position: 'top'
                },
                tooltip: {
                    mode: 'index',
                    intersect: false
                }
            }
        }
    });
}

// Load configuration on page load
async function loadConfig() {
    try {
        const response = await fetch('/api/config');
        if (response.ok) {
            const config = await response.json();
            points = config.points;
            document.getElementById('interpolation-mode').value = config.interpolationMode;
        }
    } catch (error) {
        console.error('Error loading config:', error);
    }
}

// Initialize when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
} else {
    init();
}
