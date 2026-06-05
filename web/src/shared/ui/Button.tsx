import type { ButtonHTMLAttributes, ReactNode } from 'react';

type Props = ButtonHTMLAttributes<HTMLButtonElement> & {
  icon?: ReactNode;
  tooltip?: string;
  variant?: 'primary' | 'secondary' | 'danger';
};

export function Button({ icon, children, className = '', tooltip, variant = 'secondary', type = 'button', title, ...props }: Props) {
  const button = (
    <button
      className={`app-button app-button--${variant} ${className}`}
      title={title ?? tooltip}
      type={type}
      {...props}
    >
      {icon}
      {children ? <span>{children}</span> : null}
    </button>
  );

  if (!tooltip) return button;

  return (
    <span className="app-button-wrap">
      {button}
      <span className="app-button-tooltip" role="tooltip">{tooltip}</span>
    </span>
  );
}
