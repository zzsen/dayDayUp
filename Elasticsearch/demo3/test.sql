
CREATE DATABASE IF NOT EXISTS test
DEFAULT CHARACTER SET utf8mb4
DEFAULT COLLATE utf8mb4_general_ci;

use test;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for test_role
-- ----------------------------
DROP TABLE IF EXISTS `test_role`;
CREATE TABLE `test_role`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
  `code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL COMMENT '角色code',
  `status` tinyint NULL DEFAULT NULL,
  `remark` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NULL DEFAULT NULL,
  `type` tinyint NOT NULL COMMENT '类型：1:公共角色；2:特殊角色',
  `create_time` datetime NULL DEFAULT NULL,
  `creator_id` bigint NULL DEFAULT NULL,
  `creator` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NULL DEFAULT NULL,
  `update_time` datetime NULL DEFAULT NULL,
  `modifier_id` bigint NULL DEFAULT NULL,
  `modifier` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 10 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '角色表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of test_role
-- ----------------------------
INSERT INTO `test_role` VALUES (1, '超级管理员', 'SUPBER_ADMIN', 1, '权限超级大，拥有所有权限', 2, '2021-05-27 14:09:50', 1, 'admin', '2021-05-28 10:26:28', 1, 'admin');
INSERT INTO `test_role` VALUES (6, '普通管理员', 'ADMIN', 1, '只拥有部分管理权限', 2, '2021-05-28 15:55:36', 1, 'admin', '2021-05-28 15:55:36', 1, 'admin');
INSERT INTO `test_role` VALUES (7, '公共角色', 'COMMON', 1, '所有账号基础角色', 1, '2021-07-06 15:05:47', 1, 'admin', '2021-07-06 15:05:47', 1, 'admin');
INSERT INTO `test_role` VALUES (8, '开发', 'DEV', 1, '研发人员', 0, '2021-07-09 10:46:10', 1, 'admin', '2021-07-09 10:46:10', 1, 'admin');
INSERT INTO `test_role` VALUES (9, 'mongo', 'MONGO', 1, 'mongo', 0, '2022-10-24 16:32:29', 1, 'admin', '2022-10-24 16:32:29', 1, 'admin');

SET FOREIGN_KEY_CHECKS = 1;

-- 如果是通过docker创建用户和数据库, 无须重复授权
-- 如果是通过sql创建用户和数据库, 则需要授权
GRANT ALL PRIVILEGES  ON *.* TO 'test'@'%';