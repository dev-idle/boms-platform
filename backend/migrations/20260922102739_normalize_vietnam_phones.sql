-- Rewrite every Vietnam phone into the E.164 form the API now stores, so the
-- duplicate check compares like with like. The pattern matches the API's:
-- national (0…) or international (+84 / 84, optionally followed by the national
-- 0), then a mobile number (3/5/7/8/9 + 8 digits) or a land line (2 + 9 digits).
-- Values that are not Vietnam numbers are left as they are: they still render,
-- and the next save must fix them.
UPDATE customer_profiles
SET phone = '+84' || regexp_replace(regexp_replace(phone, '[[:space:].()-]', '', 'g'), '^(0|\+?840?)', '')
WHERE regexp_replace(phone, '[[:space:].()-]', '', 'g') ~ '^(0|\+?840?)(2[0-9]{9}|[35789][0-9]{8})$';

UPDATE staff_profiles
SET phone = '+84' || regexp_replace(regexp_replace(phone, '[[:space:].()-]', '', 'g'), '^(0|\+?840?)', '')
WHERE regexp_replace(phone, '[[:space:].()-]', '', 'g') ~ '^(0|\+?840?)(2[0-9]{9}|[35789][0-9]{8})$';

UPDATE admin_profiles
SET phone = '+84' || regexp_replace(regexp_replace(phone, '[[:space:].()-]', '', 'g'), '^(0|\+?840?)', '')
WHERE regexp_replace(phone, '[[:space:].()-]', '', 'g') ~ '^(0|\+?840?)(2[0-9]{9}|[35789][0-9]{8})$';
