create table users (
  id uuid primary key,
  username text unique not null,
  password text not null
);

create table services (
  id uuid primary key,
  name text unique not null,
  logo text unique not null
);

create table accounts (
  id uuid primary key,
  user_id uuid not null,
  service_id uuid not null,
  name text not null,
  secret integer not null,
  payload text not null,
  foreign key (user_id) references users(id) on delete cascade,
  foreign key (service_id) references services(id) on delete set null
);

create table ciphers (
  id uuid primary key,
  key_value text not null
);
