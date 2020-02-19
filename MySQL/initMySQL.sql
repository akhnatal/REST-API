
USE db_sam;
CREATE TABLE IF NOT EXISTS orders
(
    id INT AUTO_INCREMENT PRIMARY KEY,
    distance INT NOT NULL,
	status CHAR(25) NOT NULL
);