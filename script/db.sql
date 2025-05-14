CREATE TABLE file_processing_logs (
    id INT IDENTITY(1,1) PRIMARY KEY,
    filename NVARCHAR(255) NOT NULL,
    status NVARCHAR(10) NOT NULL CHECK (status IN ('success', 'failed')),
    record_count INT NOT NULL DEFAULT 0,
    error_message NVARCHAR(MAX),
    created_at DATETIME2 NOT NULL,
    CONSTRAINT UQ_filename UNIQUE (filename)
);

CREATE TABLE saleorder_header (
    id INT IDENTITY(1,1) PRIMARY KEY,
    doc_no NVARCHAR(50) NOT NULL,
    on_date DATETIME2,
    delivery_date DATETIME2,
    so_customer_id NVARCHAR(50),
    customer_name NVARCHAR(255),
    status NVARCHAR(50),
    territory_code NVARCHAR(50),
    total_amount DECIMAL(18, 2),
    total_vat DECIMAL(18, 2),
    remark NVARCHAR(MAX),
    source_file NVARCHAR(255) NOT NULL,
    created_at DATETIME2 NOT NULL
);

CREATE TABLE saleorder_item (
    id INT IDENTITY(1,1) PRIMARY KEY,
    doc_no NVARCHAR(50) NOT NULL,
    item_id NVARCHAR(50),
    product_code NVARCHAR(50),
    so_product_id NVARCHAR(50),
    sku_unit_type_id NVARCHAR(50),
    quantity DECIMAL(18, 2),
    price DECIMAL(18, 2),
    amount DECIMAL(18, 2),
    vat DECIMAL(18, 2),
    vat_rate DECIMAL(18, 2),
    item_type NVARCHAR(50),
    order_rank INT,
    ref_item_id NVARCHAR(50),
    io_number NVARCHAR(50),
    source_file NVARCHAR(255) NOT NULL,
    created_at DATETIME2 NOT NULL
);

CREATE TABLE saleorder_summary (
    id INT IDENTITY(1,1) PRIMARY KEY,
    header_count INT,
    item_count INT,
    source_file NVARCHAR(255) NOT NULL,
    created_at DATETIME2 NOT NULL
);