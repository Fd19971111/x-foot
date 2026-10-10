-- 客户管理模块：客户信息补充字段
--
-- 背景：site_customer 原有 customer_name（客户名称）承载「姓名」，
-- 但业务上需要把「公司」与「身份证号」分开保存，因此补充两列。
--
-- 执行方式（幂等）：本文件由 api/hack/dbprobe 中的迁移程序读取执行，
-- 程序会先检查列是否存在，已存在则跳过，可重复运行。

ALTER TABLE `site_customer`
  ADD COLUMN `company_name` varchar(100) NULL DEFAULT NULL COMMENT '公司' AFTER `customer_name`,
  ADD COLUMN `id_card` varchar(20) NULL DEFAULT NULL COMMENT '身份证号' AFTER `phone`;
