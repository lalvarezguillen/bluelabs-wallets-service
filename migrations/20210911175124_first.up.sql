CREATE TABLE public.wallets (
	id bigserial NOT NULL,
	created_at timestamptz default current_timestamp,
	name text,
	balance int8 DEFAULT 0,
	CONSTRAINT wallets_pkey PRIMARY KEY (id)
);

CREATE TABLE public.balance_changes (
	id bigserial NOT NULL,
	created_at timestamptz default current_timestamp,
	amount int8,
	operation text,
	balance_before int8,
	balance_after int8,
	reference text NULL,
	wallet_id int8,
	CONSTRAINT balance_changes_pkey PRIMARY KEY (id),
	CONSTRAINT fk_balance_changes_wallet FOREIGN KEY (wallet_id) REFERENCES wallets(id)
);