import type { ReactNode } from 'react';

type InfoTileProps = {
  title: string;
  detail?: ReactNode;
  badge?: ReactNode;
  icon?: ReactNode;
  className?: string;
  titleClassName?: string;
};

export function InfoTile({ title, detail, badge, icon, className = '', titleClassName = '' }: InfoTileProps) {
  return (
    <div className={`info-tile ${className}`}>
      <div className="info-tile__header">
        {icon ? <div className="info-tile__icon">{icon}</div> : null}
        <div className="min-w-0">
          <div className={`info-tile__title text-wrap-safe ${titleClassName}`} title={title}>{title}</div>
          {detail ? <div className="info-tile__detail text-wrap-safe">{detail}</div> : null}
        </div>
      </div>
      {badge ? <div className="min-w-0">{badge}</div> : null}
    </div>
  );
}
