import type { FanCurve } from './types';

export class FanCurveManager {
  private curves: Map<string, FanCurve> = new Map();

  addCurve(curve: FanCurve): void {
    this.curves.set(curve.id, curve);
  }

  getCurve(id: string): FanCurve | undefined {
    return this.curves.get(id);
  }

  getAllCurves(): FanCurve[] {
    return Array.from(this.curves.values());
  }

  updateCurve(id: string, curve: FanCurve): void {
    this.curves.set(id, curve);
  }

  deleteCurve(id: string): void {
    this.curves.delete(id);
  }

  // Calculate fan speed for a given temperature using linear interpolation
  calculateFanSpeed(curve: FanCurve, temperature: number): number {
    const points = [...curve.points].sort((a, b) => a.temperature - b.temperature);
    
    if (points.length === 0) return 0;
    if (points.length === 1) return points[0].fanSpeed;
    
    // If temperature is below or above all points, use the closest point
    if (temperature <= points[0].temperature) return points[0].fanSpeed;
    if (temperature >= points[points.length - 1].temperature) return points[points.length - 1].fanSpeed;
    
    // Find the two points to interpolate between
    for (let i = 0; i < points.length - 1; i++) {
      const p1 = points[i];
      const p2 = points[i + 1];
      
      if (temperature >= p1.temperature && temperature <= p2.temperature) {
        // Linear interpolation
        const ratio = (temperature - p1.temperature) / (p2.temperature - p1.temperature);
        return p1.fanSpeed + ratio * (p2.fanSpeed - p1.fanSpeed);
      }
    }
    
    return 0;
  }
}
