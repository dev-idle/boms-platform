import { AuthFormShell } from "./auth-form-shell";

type AuthFormSkeletonProps = {
  description: string;
  fields?: number;
  title: string;
};

/** Auth form placeholder: the real heading ships in the static shell, fields stream in. */
export function AuthFormSkeleton({ description, fields = 2, title }: AuthFormSkeletonProps) {
  return (
    <AuthFormShell description={description} title={title}>
      <div aria-busy="true" aria-label="Loading form" className="auth-form" role="status">
        {Array.from({ length: fields }, (_, index) => (
          <div className="field-control" key={index}>
            <div className="skeleton h-3 w-24" />
            <div className="skeleton h-(--field-control-height) w-full" />
          </div>
        ))}
        <div className="skeleton h-(--btn-control-min-height) w-full" />
      </div>
    </AuthFormShell>
  );
}
