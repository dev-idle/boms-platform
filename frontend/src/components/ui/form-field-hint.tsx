type FormFieldHintProps = {
  children: string;
  id?: string;
};

/** Muted helper between label and control — copy from `FORM_FIELD_HINT`. */
export function FormFieldHint({ children, id }: FormFieldHintProps) {
  return (
    <p className="form-field-hint" id={id}>
      {children}
    </p>
  );
}
