alter table t_user add column current_balance numeric(10, 2) not null default 0;
alter table t_user add column withdrawn_balance numeric(10, 2) not null default 0;