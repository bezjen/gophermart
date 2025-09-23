alter table t_user
drop column if exists current_balance,
drop column if exists withdrawn_balance;