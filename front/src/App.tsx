import { useState } from 'react';
import { Line } from 'react-chartjs-2';
import {
    Chart as ChartJS,
    CategoryScale,
    LinearScale,
    PointElement,
    LineElement,
    Title,
    Tooltip,
    Legend,
    Filler
} from 'chart.js';


export interface FanCurvePoint {
    temperature: number; // 0-100
    fanSpeed: number; // percentage (0-100) or RPM value
}

ChartJS.register(
    CategoryScale,
    LinearScale,
    PointElement,
    LineElement,
    Title,
    Tooltip,
    Legend,
    Filler
);

type InterpolationMode = 'gradual' | 'hardcut';

function App() {
    const [points, setPoints] = useState<FanCurvePoint[]>([
        { temperature: 20, fanSpeed: 20 },
        { temperature: 80, fanSpeed: 100 }
    ]);
    const [interpolationMode, setInterpolationMode] = useState<InterpolationMode>('gradual');

    const calculateFanSpeedGradual = (temperature: number): number => {
        const sortedPoints = [...points].sort((a, b) => a.temperature - b.temperature);

        if (sortedPoints.length === 0) return 0;
        if (sortedPoints.length === 1) return sortedPoints[0].fanSpeed;

        if (temperature <= sortedPoints[0].temperature) return sortedPoints[0].fanSpeed;
        if (temperature >= sortedPoints[sortedPoints.length - 1].temperature)
            return sortedPoints[sortedPoints.length - 1].fanSpeed;

        for (let i = 0; i < sortedPoints.length - 1; i++) {
            const p1 = sortedPoints[i];
            const p2 = sortedPoints[i + 1];

            if (temperature >= p1.temperature && temperature <= p2.temperature) {
                const ratio = (temperature - p1.temperature) / (p2.temperature - p1.temperature);
                return p1.fanSpeed + ratio * (p2.fanSpeed - p1.fanSpeed);
            }
        }

        return 0;
    };

    const calculateFanSpeedHardCut = (temperature: number): number => {
        const sortedPoints = [...points].sort((a, b) => a.temperature - b.temperature);

        if (sortedPoints.length === 0) return 0;
        if (sortedPoints.length === 1) return sortedPoints[0].fanSpeed;

        if (temperature < sortedPoints[0].temperature) return sortedPoints[0].fanSpeed;

        for (let i = sortedPoints.length - 1; i >= 0; i--) {
            if (temperature >= sortedPoints[i].temperature) {
                return sortedPoints[i].fanSpeed;
            }
        }

        return sortedPoints[0].fanSpeed;
    };

    const calculateFanSpeed = (temperature: number): number => {
        if (interpolationMode === 'gradual') {
            return calculateFanSpeedGradual(temperature);
        } else {
            return calculateFanSpeedHardCut(temperature);
        }
    };

    const generateChartData = () => {
        const sortedPoints = [...points].sort((a, b) => a.temperature - b.temperature);
        const curveData: { x: number; y: number }[] = [];

        for (let temp = 0; temp <= 100; temp += 1) {
            curveData.push({ x: temp, y: calculateFanSpeed(temp) });
        }

        return {
            datasets: [
                {
                    label: 'Fan Curve',
                    data: curveData,
                    borderColor: 'rgb(59, 130, 246)',
                    backgroundColor: 'rgba(59, 130, 246, 0.1)',
                    borderWidth: 2,
                    fill: true,
                    tension: 0,
                    pointRadius: 0,
                },
                {
                    label: 'Control Points',
                    data: sortedPoints.map(p => ({ x: p.temperature, y: p.fanSpeed })),
                    borderColor: 'rgb(239, 68, 68)',
                    backgroundColor: 'rgb(239, 68, 68)',
                    borderWidth: 0,
                    pointRadius: 6,
                    pointHoverRadius: 8,
                    showLine: false,
                }
            ]
        };
    };

    const chartOptions = {
        responsive: true,
        maintainAspectRatio: true,
        aspectRatio: 2,
        scales: {
            x: {
                type: 'linear' as const,
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
    };

    const addPoint = () => {
        setPoints([...points, { temperature: 50, fanSpeed: 50 }]);
    };

    const updatePoint = (index: number, field: 'temperature' | 'fanSpeed', value: number) => {
        const newPoints = [...points];
        newPoints[index][field] = value;
        setPoints(newPoints);
    };

    const removePoint = (index: number) => {
        setPoints(points.filter((_, i) => i !== index));
    };

    return (
        <div className="bg-gray-100 min-h-screen">
            <div className="container mx-auto px-4 py-8 max-w-5xl">
                <h1 className="text-3xl font-bold mb-8 text-gray-800">Fan Curve</h1>

                {/* Interpolation Mode */}
                <div className="bg-white rounded-lg shadow-md p-6 mb-6">
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                        Interpolation Mode
                    </label>
                    <select
                        value={interpolationMode}
                        onChange={(e) => setInterpolationMode(e.target.value as InterpolationMode)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                        <option value="gradual">Gradual (Linear)</option>
                        <option value="hardcut">Hard Cuts (Step)</option>
                    </select>
                </div>

                {/* Graph */}
                <div className="bg-white rounded-lg shadow-md p-6 mb-6">
                    <Line data={generateChartData()} options={chartOptions} />
                </div>

                {/* Points */}
                <div className="bg-white rounded-lg shadow-md p-6">
                    <h2 className="text-xl font-semibold mb-4">Curve Points</h2>
                    <div className="space-y-2">
                        {points.map((point, index) => (
                            <div key={index} className="grid grid-cols-3 gap-2 items-center">
                                <input
                                    type="number"
                                    min="0"
                                    max="100"
                                    value={point.temperature}
                                    onChange={(e) => updatePoint(index, 'temperature', parseFloat(e.target.value) || 0)}
                                    placeholder="Temp (°C)"
                                    className="px-3 py-2 border border-gray-300 rounded-md"
                                />
                                <input
                                    type="number"
                                    min="0"
                                    max="100"
                                    value={point.fanSpeed}
                                    onChange={(e) => updatePoint(index, 'fanSpeed', parseFloat(e.target.value) || 0)}
                                    placeholder="Speed (%)"
                                    className="px-3 py-2 border border-gray-300 rounded-md"
                                />
                                <button
                                    onClick={() => removePoint(index)}
                                    className="bg-red-500 hover:bg-red-600 text-white px-3 py-2 rounded-md text-sm"
                                >
                                    Remove
                                </button>
                            </div>
                        ))}
                    </div>
                    <button
                        onClick={addPoint}
                        className="mt-4 bg-green-500 hover:bg-green-600 text-white px-4 py-2 rounded-md"
                    >
                        Add Point
                    </button>
                </div>
            </div>
        </div>
    );
}

export default App;
