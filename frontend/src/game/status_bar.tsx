import type { PlayerStatus } from './socket_message';

const STATUS_BAR_COLOR_RED = '#ff0000';
const STATUS_BAR_COLOR_BLUE = '#5fcde4';

type StatusBarProps = {
  status: PlayerStatus | null;
}

export const StatusBar = ({ status }: StatusBarProps) => {
  if (status === null) {
    return null;
  }

  return (
    <div className="status-bar">
      <StatusMeter label="HP" current={status.hp} max={status.maxHp} color={STATUS_BAR_COLOR_RED} />
      <StatusMeter label="MP" current={status.mp} max={status.maxMp} color={STATUS_BAR_COLOR_BLUE} />
    </div>
  )
};

type StatusMeterProps = {
  label: string;
  current: number;
  max: number;
  color: string;
};

const StatusMeter = ({ label, current, max, color }: StatusMeterProps) => {
  const fillPercent = max > 0 ? Math.min(Math.max(current / max, 0), 1) * 100 : 0;

  return (
    <div className="status-meter">
      <span className="status-meter-text">{label} {current} / {max}</span>
      <div className="status-meter-track">
        <div className="status-meter-fill" style={{ width: `${fillPercent}%`, backgroundColor: color }}></div>
      </div>
    </div>
  );
};
