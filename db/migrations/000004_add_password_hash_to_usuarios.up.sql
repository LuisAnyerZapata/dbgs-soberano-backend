ALTER TABLE dbgs_schema.usuarios
ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);

UPDATE dbgs_schema.usuarios
SET password_hash = '$2a$10$ml5UjdnV8l12NY8ccG4qXexZ10hgmnEI041G0lGQ.MOaBqfrF.SAK'
WHERE username = 'admin' AND (password_hash IS NULL OR password_hash = '');

UPDATE dbgs_schema.usuarios
SET password_hash = '$2a$10$liR7AbaN5dtBpROpoh5kHercsCZbhYiz.wic4tB8txOAj0OOkxF2.'
WHERE username = 'dba' AND (password_hash IS NULL OR password_hash = '');

UPDATE dbgs_schema.usuarios
SET password_hash = '$2a$10$LxWusq0ZswovCRRzZ/Y5FOcXi1C9q7542vHlywlZMXO55enCWvF4S'
WHERE username = 'dev' AND (password_hash IS NULL OR password_hash = '');

UPDATE dbgs_schema.usuarios
SET password_hash = '$2a$10$arx2Qq.1mIJEJyNb3KuGiOGncuIywUUfCO1W71ULnSV2xyp30Ejcy'
WHERE username = 'auditor' AND (password_hash IS NULL OR password_hash = '');

UPDATE dbgs_schema.usuarios
SET password_hash = '$2a$10$.7QESBL97RRTPxz/am09P.wjRKt6CJQYAD0VbehtyUwd3ekCwz0pC'
WHERE username = 'service_api' AND (password_hash IS NULL OR password_hash = '');
