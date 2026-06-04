import type { ButtonHTMLAttributes, ReactNode } from 'react';

type Props = ButtonHTMLAttributes<HTMLButtonElement> & {
  icon?: ReactNode;
  variant?: 'primary' | 'secondary' | 'danger';
};

export function Button({ icon, children, className = '', variant = 'secondary', type = 'button', ...props }: Props) {
  return (
    <button
      className={`app-button app-button--${variant} ${className}`}
      type={type}
      {...props}
    >
      {icon}
      {children ? <span>{children}</span> : null}
    </button>
  );
}
