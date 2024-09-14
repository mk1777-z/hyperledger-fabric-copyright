# login

1.用户需要输入username和password（用这个命名变量）（***必须使用真实姓名***）

2.用JavaScript将密码加密传输给后端（这个需要做接口，路由为“/login”）

3判断用户输入是否完整，是否有违规字符（防止SQL注入）

3.后端返回200或400或500{

200：登录成功，后端生成一个token保存在前端作为区别用户信息

400：用户密码错误

500：服务器出现问题

}

4.登录成功要跳转到display.html

# register

1.输入同上

2.判断同上

3.后端会判断是否有这个用户返回{

200：注册成功

400：用户已存在

500：服务器错误

}

# 版权管理

1.从后端生成项目信息发送到前端，生成格式为json，

{

 ‘item1’:{

   'name':xxx

    'owner':xxxx

   'price':2000

   'simple_dsc':xxxxxxxxxxxx

‘id’:xxxxx

   'img' :base64编码

 }

item2’:{

   'name':xxx

   'owner':xxxx

   'price':2000

   'simple_dsc':xxxxxxxxxxxx

   'id':xxx

    'img' :base64编码

}

}

2.查看详情接口

跳转到information.html?id=xxx 路由增加请求参数id

# information

1.直接由路由获得项目id然后后端返回json：

{

'name':xxx

'owner':xxxx

'price':2000

'simple_dsc':xxxxxxxxxxxx

id’:xxxxx

'img' :base64编码

'dsc':xxxxxxxxxxxx

'start_time' :xxxx/xx/xx

'trace':unknown

'on_sale':xxxx***（如果处于不可售卖，购买按钮应该是禁止选中的）***

}

2.点击购买之后跳出弹窗contract，用户需要输入自己的真实姓名，确认。

3.确认之后生成pdf，发出post请求，带有token作为用户区分办法，然后修改世界状态,修改数据库owner。

4跳转到display.html

# homepage

1.发出post请求，请求头包含token，返回用户各个信息

{

‘item1’:{

   'name':xxx

   'owner':xxxx

   'price':2000

   'simple_dsc':xxxxxxxxxxxx

   'id':xxx

    'img' :base64编码

    'on_sale':xxx

}

‘item2’:{

   'name':xxx

   'owner':xxxx

   'price':2000

   'simple_dsc':xxxxxxxxxxxx

   'id':xxx

    'img' :base64编码

    'on_sale':xxxx

}

'username':xxxx

}

2.修改操作：修改name，price，simple_dsc，img，dsc，是否继续售卖

3.项目可以查看information，路由同上

# Upload

1.返回后端json

{

'name':xxx

'simple_dsc':xxx

'dsc':xxx

'price':xxx

'img':base64

}

2.返回创建成功代码

{

同上

}
