## ADDED Requirements

### Requirement: 动态插件通过命名通知通道发送宿主通知

系统 SHALL 为动态插件提供受治理的通知服务，插件只能通过宿主授权的通知通道发送站内信、邮件、Webhook 等通知。

#### Scenario: 插件使用授权通知通道

- **WHEN** 插件调用通知服务向已授权的`host-notify-channel`发送通知
- **THEN** 宿主校验通道权限、模板或消息体约束
- **AND** 宿主按对应通知通道完成发送

#### Scenario: 插件尝试使用未授权通知通道

- **WHEN** 插件调用一个未授权的通知通道
- **THEN** 宿主拒绝该调用
- **AND** 宿主不向 guest 暴露宿主通知后端实现细节
