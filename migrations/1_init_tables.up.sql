create table users (
  id uuid primary key,
  username text unique not null,
  password text not null
);

create table services_creds (
  id uuid primary key,
  user_id uuid not null,
  name text not null,
  secret integer not null,
  payload text not null,
  foreign key (user_id) references users(id) on delete cascade
);
