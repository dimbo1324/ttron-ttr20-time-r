import type { ButtonHTMLAttributes, ReactNode } from 'react';

type Props = ButtonHTMLAttributes<HTMLButtonElement> & {
  icon?: ReactNode;
  variant?: 'primary' | 'secondary' | 'danger';
};

export function Button({ icon, children, className = '', variant = 'secondary', ...props }: Props) {
  return (
    <button
      className={`app-button app-button--${variant} ${className}`}
      {...props}
    >
      {icon}
      <span>{children}</span>
    </button>
  );
}
