CREATE TABLE `paste` (
  `id` varchar(36) NOT NULL,
  `timestamp` int(11) NOT NULL,
  `expires` int(11) NOT NULL,
  `content` longtext NOT NULL,
  `syntax` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


ALTER TABLE `paste`
  ADD PRIMARY KEY (`id`);
