import { PASSWORD_REQUIREMENT_CHECKS } from "@/lib/validation/password";
import { cn } from "@/lib/utils";

type AuthPasswordChecklistProps = {
  password: string;
};

export function AuthPasswordChecklist({ password }: AuthPasswordChecklistProps) {
  return (
    <ul aria-label="Password requirements" className="auth-password-checklist">
      {PASSWORD_REQUIREMENT_CHECKS.map((check) => {
        const ok = check.test(password);

        return (
          <li
            key={check.id}
            className={cn(
              "auth-password-checklist-item",
              ok && "auth-password-checklist-item-met",
            )}
          >
            <span aria-hidden="true" className="auth-password-checklist-dot" />
            {check.label}
          </li>
        );
      })}
    </ul>
  );
}
