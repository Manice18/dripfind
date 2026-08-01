type IconProps = {
  className?: string;
  title?: string;
};

export function IconUser({ className, title }: IconProps) {
  return (
    <svg
      className={className}
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.75"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden={title ? undefined : true}
      role={title ? "img" : undefined}
    >
      {title ? <title>{title}</title> : null}
      <circle cx="12" cy="8" r="3.25" />
      <path d="M5.5 19.25c1.4-3.1 3.7-4.75 6.5-4.75s5.1 1.65 6.5 4.75" />
    </svg>
  );
}

export function IconLogout({ className, title }: IconProps) {
  return (
    <svg
      className={className}
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.75"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden={title ? undefined : true}
      role={title ? "img" : undefined}
    >
      {title ? <title>{title}</title> : null}
      <path d="M10 4.5H6.75A2.25 2.25 0 0 0 4.5 6.75v10.5A2.25 2.25 0 0 0 6.75 19.5H10" />
      <path d="M13.5 15.5 17.5 12l-4-3.5" />
      <path d="M17.25 12H9" />
    </svg>
  );
}
