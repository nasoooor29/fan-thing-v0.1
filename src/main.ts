import { Chart } from 'chart.js/auto';
import type { FanCurvePoint } from './types';

type InterpolationMode = 'gradual' | 'hardcut';

let points: FanCurvePoint[] = [];
let chart: Chart;
let interpolationMode: InterpolationMode = 'gradual';

function calculateFanSpeedGradual(temperature: number): number {
  const sortedPoints = [...points].sort((a, b) => a.temperature - b.temperature);
  
  if (sortedPoints.length === 0) return 0;
  if (sortedPoints.length === 1) return sortedPoints[0].fanSpeed;
  
  if (temperature <= sortedPoints[0].temperature) return sortedPoints[0].fanSpeed;
  if (temperature >= sortedPoints[sortedPoints.length - 1].temperature) return sortedPoints[sortedPoints.length - 1].fanSpeed;
  
  for (let i = 0; i < sortedPoints.length - 1; i++) {
    const p1 = sortedPoints[i];
    const p2 = sortedPoints[i + 1];
    
    if (temperature >= p1.temperature && temperature <= p2.temperature) {
      const ratio = (temperature - p1.temperature) / (p2.temperature - p1.temperature);
      return p1.fanSpeed + ratio * (p2.fanSpeed - p1.fanSpeed);
    }
  }
  
  return 0;
}

function calculateFanSpeedHardCut(temperature: number): number {
  const sortedPoints = [...points].sort((a, b) => a.temperature - b.temperature);
  
  if (sortedPoints.length === 0) return 0;
  if (sortedPoints.length === 1) return sortedPoints[0].fanSpeed;
  
  if (temperature < sortedPoints[0].temperature) return sortedPoints[0].fanSpeed;
  
  // Find the last point where temperature >= point.temperature
  for (let i = sortedPoints.length - 1; i >= 0; i--) {
    if (temperature >= sortedPoints[i].temperature) {
      return sortedPoints[i].fanSpeed;
    }
  }
  
  return sortedPoints[0].fanSpeed;
}

function calculateFanSpeed(temperature: number): number {
  if (interpolationMode === 'gradual') {
    return calculateFanSpeedGradual(temperature);
  } else {
    return calculateFanSpeedHardCut(temperature);
  }
}

function updateChart(): void {
  const sortedPoints = [...points].sort((a, b) => a.temperature - b.temperature);
  
  // Generate curve data
  const curveData: { x: number; y: number }[] = [];
  for (let temp = 0; temp <= 100; temp += 1) {
    curveData.push({ x: temp, y: calculateFanSpeed(temp) });
  }
  
  chart.data.datasets[0].data = curveData;
  chart.data.datasets[1].data = sortedPoints.map(p => ({ x: p.temperature, y: p.fanSpeed }));
  chart.update();
}

function renderPoints(): void {
  const pointsList = document.getElementById('pointsList');
  if (!pointsList) return;
  
  pointsList.innerHTML = '';
  
  points.forEach((point, index) => {
    const pointDiv = document.createElement('div');
    pointDiv.className = 'grid grid-cols-3 gap-2 items-center';
    
    pointDiv.innerHTML = `
      <input type="number" class="temp-input px-3 py-2 border border-gray-300 rounded-md" 
             min="0" max="100" value="${point.temperature}" placeholder="Temp (°C)">
      <input type="number" class="speed-input px-3 py-2 border border-gray-300 rounded-md" 
             min="0" max="100" value="${point.fanSpeed}" placeholder="Speed (%)">
      <button class="remove-btn bg-red-500 hover:bg-red-600 text-white px-3 py-2 rounded-md text-sm">
        Remove
      </button>
    `;
    
    const tempInput = pointDiv.querySelector('.temp-input') as HTMLInputElement;
    const speedInput = pointDiv.querySelector('.speed-input') as HTMLInputElement;
    const removeBtn = pointDiv.querySelector('.remove-btn');
    
    tempInput.addEventListener('input', () => {
      point.temperature = parseFloat(tempInput.value) || 0;
      updateChart();
    });
    
    speedInput.addEventListener('input', () => {
      point.fanSpeed = parseFloat(speedInput.value) || 0;
      updateChart();
    });
    
    removeBtn?.addEventListener('click', () => {
      points.splice(index, 1);
      renderPoints();
      updateChart();
    });
    
    pointsList.appendChild(pointDiv);
  });
}

document.addEventListener('DOMContentLoaded', () => {
  const canvas = document.getElementById('fanChart') as HTMLCanvasElement;
  const addPointBtn = document.getElementById('addPoint');
  const interpolationModeSelect = document.getElementById('interpolationMode') as HTMLSelectElement;
  
  // Initialize chart
  chart = new Chart(canvas, {
    type: 'line',
    data: {
      datasets: [
        {
          label: 'Fan Curve',
          data: [],
          borderColor: 'rgb(59, 130, 246)',
          backgroundColor: 'rgba(59, 130, 246, 0.1)',
          borderWidth: 2,
          fill: true,
          tension: 0,
          pointRadius: 0,
          stepped: false
        },
        {
          label: 'Control Points',
          data: [],
          borderColor: 'rgb(239, 68, 68)',
          backgroundColor: 'rgb(239, 68, 68)',
          borderWidth: 0,
          pointRadius: 6,
          pointHoverRadius: 8,
          showLine: false
        }
      ]
    },
    options: {
      responsive: true,
      maintainAspectRatio: true,
      aspectRatio: 2,
      scales: {
        x: {
          type: 'linear',
          min: 0,
          max: 100,
          title: {
            display: true,
            text: 'Temperature (°C)'
          }
        },
        y: {
          min: 0,
          max: 100,
          title: {
            display: true,
            text: 'Fan Speed (%)'
          }
        }
      }
    }
  });
  
  interpolationModeSelect.addEventListener('change', () => {
    interpolationMode = interpolationModeSelect.value as InterpolationMode;
    updateChart();
  });
  
  addPointBtn?.addEventListener('click', () => {
    points.push({ temperature: 50, fanSpeed: 50 });
    renderPoints();
    updateChart();
  });
  
  // Add initial points
  points = [
    { temperature: 20, fanSpeed: 20 },
    { temperature: 80, fanSpeed: 100 }
  ];
  
  renderPoints();
  updateChart();
});

