import './status_bar.css';
import type { PlayerStatus } from './socket_message';
import { TextColor, TEXT_COLOR_VALUES } from './terminal/text_color';

type StatusBarProps = {
  status: PlayerStatus | null;
};

export const StatusBar = ({ status }: StatusBarProps) => {
  if (status === null) {
    return null;
  }

  return (
    <div className="status-bar">
      <StatusMeter label="HP" current={status.health} max={status.maxHealth} color={TEXT_COLOR_VALUES[TextColor.RED]} />
      <StatusMeter label="MP" current={status.mana} max={status.maxMana} color={TEXT_COLOR_VALUES[TextColor.BLUE]} />
    </div>
  );
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
