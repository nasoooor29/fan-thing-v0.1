export type FanControlMode = 'percentage' | 'rpm';

export interface FanCurvePoint {
  temperature: number; // 0-100
  fanSpeed: number; // percentage (0-100) or RPM value
}

export interface FanCurve {
  id: string;
  name: string;
  mode: FanControlMode;
  points: FanCurvePoint[];
}
