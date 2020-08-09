create table if not exists shareholder (
    id serial primary key,
    name text not null,
    email text not null,
    email_verified bool not null default false,
    mobile text not null,
    mobile_verified bool not null default false,
    folio_number text not null,
    certificate_number text not null,
    pan_number text not null,
    agree_terms bool not null
);
