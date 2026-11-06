import type { PlayerStatus } from './socket_message';

const STATUS_BAR_COLOR_RED = '#ff0000';
const STATUS_BAR_COLOR_BLUE = '#5fcde4';

type StatusBarProps = {
  status: PlayerStatus | null;
}

// With no status (the player isn't in the world), the meters show only their labels and empty tracks
export const StatusBar = ({ status }: StatusBarProps) => {
  return (
    <div className="status-bar">
      <StatusMeter label="HP" current={status?.hp ?? null} max={status?.maxHp ?? null} color={STATUS_BAR_COLOR_RED} />
      <StatusMeter label="MP" current={status?.mp ?? null} max={status?.maxMp ?? null} color={STATUS_BAR_COLOR_BLUE} />
    </div>
  )
};

type StatusMeterProps = {
  label: string;
  current: number | null;
  max: number | null;
  color: string;
};

const StatusMeter = ({ label, current, max, color }: StatusMeterProps) => {
  const hasValues = current !== null && max !== null;
  const fillPercent = hasValues && max > 0 ? Math.min(Math.max(current / max, 0), 1) * 100 : 0;

  return (
    <div className="status-meter">
      <span className="status-meter-text">{hasValues ? `${label} ${current} / ${max}` : label}</span>
      <div className="status-meter-track">
        <div className="status-meter-fill" style={{ width: `${fillPercent}%`, backgroundColor: color }}></div>
      </div>
    </div>
  );
};
