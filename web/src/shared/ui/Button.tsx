import type { ButtonHTMLAttributes, ReactNode } from 'react';

type Props = ButtonHTMLAttributes<HTMLButtonElement> & {
  icon?: ReactNode;
  variant?: 'primary' | 'secondary' | 'danger';
};

export function Button({ icon, children, className = '', variant = 'secondary', ...props }: Props) {
  const variants = {
    primary: 'border-signal/50 bg-signal/15 text-signal hover:bg-signal/25',
    secondary: 'border-line bg-white/5 text-zinc-200 hover:bg-white/10',
    danger: 'border-fault/50 bg-fault/10 text-fault hover:bg-fault/20',
  };
  return (
    <button
      className={`inline-flex min-h-9 items-center justify-center gap-2 rounded-md border px-3 py-1.5 text-sm font-medium transition disabled:cursor-not-allowed disabled:opacity-50 ${variants[variant]} ${className}`}
      {...props}
    >
      {icon}
      <span>{children}</span>
    </button>
  );
}
