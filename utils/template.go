package utils

const ActiveContentType = "text/html; charset=UTF-8"
const ActiveFrom = "Dockboard <no-reply@dockboard.org>"
const ActiveSubject = "Dockboard 激活邮件"
const ActiveBodyTemplate = `
<table border="0" width="100%" cellpadding="0" cellspacing="0" bgcolor="#E6E5E7">
  <tr>
    <td height="50"></td>
    </tr>
    
    <!-- ======= header ======= -->
    <tr>
      <td align="center">
        <table border="0" align="center" width="590" cellpadding="0" cellspacing="0" bgcolor="#FFFFFF" class="container590">
          <tr>
            <td align="center">
              <table border="0" align="center" width="590" cellpadding="0" cellspacing="0" class="container590">
                <tr>
                  <td>
                    <table border="0" align="left" cellpadding="0" cellspacing="0" width="190" bgcolor="#2C3E50" class="logo">
                      <tr>
                        <td height="25" style="font-size: 25px; line-height: 25px;">&nbsp;</td>
                      </tr>
                      <tr>
                        <td align="center">
                          <table border="0" cellpadding="0" cellspacing="0">
                            <tr>
                              <td align="center"><a href="" style="display: block; border-style: none !important; border: 0 !important;"><img width="194" height="42" border="0" style="display: block; width: 194px; height: 42px;" src="http://dockboard.qiniudn.com/dockboard_email_logo_194x42.png" alt="https://www.dockboard.org" /></a></td>
                            </tr>
                          </table>
                        </td>
                      </tr>
                      <tr>
                        <td height="25" style="font-size: 25px; line-height: 25px;">&nbsp;</td>
                      </tr>
                    </table>
                    <table border="0" align="left" cellpadding="0" cellspacing="0" class="hideforiphone">
                      <tr>
                        <td height="20" width="20" style="font-size: 20px; line-height: 20px;">&nbsp;</td>
                      </tr>
                    </table>
                    <table border="0" align="right" cellpadding="0" cellspacing="0" class="date">
                      <tr>
                        <td align="center" valign="middle">
                          <table width="300" border="0" cellpadding="0" cellspacing="0" align="center" class="date-inside">
                            <tr>
                              <td height="30" style="font-size: 30px; line-height: 40px;">&nbsp;</td>
                            </tr>
                            <tr>
                              <td align="right" style="color: #959da6; font-size: 16px; font-weight: normal; font-family:'Source Sans Pro' Arial, sans-serif;">Dockboard Hub / Build / CI / CD&nbsp;&nbsp;&nbsp;</td>
                            </tr>
                            <tr>
                              <td height="30" style="font-size: 30px; line-height: 30px;">&nbsp;</td>
                            </tr>                                                        
                          </table>
                        </td>
                      </tr>
                    </table>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
        </table>
      </td>
    </tr>
    <!-- ======= end header ======= -->

    <tr>
      <td height="30" style="font-size: 30px; line-height: 30px;">&nbsp;</td>
    </tr>

    <!-- ======= Main section ======= -->

    <!-- ======= CTA ======= -->
    <tr>
      <td align="center">
        <table border="0" align="center" width="590" cellpadding="0" cellspacing="0" bgcolor="#FFFFFF" style="border-collapse:collapse; mso-table-lspace:0pt; mso-table-rspace:0pt;" class="container590">
          <tr>
            <td height="45" style="font-size: 45px; line-height: 45px;">&nbsp;</td>
              </tr>
                <tr>
                  <td align="center" style="color: #222222; font-size: 24px; font-family: 'Ubuntu', Arial, sans-serif; mso-line-height-rule: exactly;" class="cta-header">
                    <div>
                      请激活您的 Dockboard 帐户
                    </div>
                  </td>
                </tr>
                <tr>
                  <td height="25" style="font-size: 25px; line-height: 25px;">&nbsp;</td>
                </tr>
                <tr>
                  <td>
                    <table border="0" align="center" width="490" cellpadding="0" cellspacing="0" bgcolor="#FFFFFF" class="container580">
                      <tr>
                        <td align="center" style="color: #adb3ba; font-size: 16px; font-weight: 300; font-family: 'Source Sans Pro', Arial, sans-serif; mso-line-height-rule: exactly; line-height: 24px;">
                          <div style="line-height: 24px;">
                            感谢您注册 Dockboard 帐户！请点击下面的激活按钮激活帐号并登录到控制台设置个人信息。
                          </div>
                        </td>
                      </tr>
                    </table>
                  </td>
                </tr>
                <tr>
                  <td height="40" style="font-size: 40px; line-height: 40px;">&nbsp;</td>
                </tr>
                <tr>
                  <td align="center"><a href="{{.ActiveLink}}" style="display: block; width: 140px; height: 40px; border-style: none !important; border: 0 !important;"><img width="140" height="40" border="0" style="display: block; width: 140px; height: 40px;" src="http://dockboard.qiniudn.com/dockboard_email_active-btn.png" alt="激活帐号" /></a></td>
                </tr>
                <tr>
                  <td height="45" style="font-size: 45px; line-height: 45px;">&nbsp;</td>
                </tr>
        </table>
      </td>
    </tr>
    <!-- ======= end CTA ======= -->

    <tr>
      <td height="10" style="font-size: 10px; line-height: 10px;">&nbsp;</td>
    </tr>

    <!-- ======= footer ======= -->
    <tr>
      <td align="center">
        <table border="0" align="center" width="590" cellpadding="0" cellspacing="0" bgcolor="#FFFFFF" class="container590">
          <tr>
            <td align="center">
              <table border="0" align="center" width="560" cellpadding="0" cellspacing="0" class="container580">
                <tr>
                  <td height="20" style="font-size: 20px; line-height: 20px;">&nbsp;</td>
                </tr>
                <tr>
                  <td></td>
                  <td align="center">
                    <table border="0" align="left" cellpadding="0" cellspacing="0" style="border-collapse:collapse; mso-table-lspace:0pt; mso-table-rspace:0pt;" class="container580">
                      <tr>
                        <td align="center" style="color: #adb3ba; font-size: 13px; font-family: 'Source Sans pro', Arial, sans-serif; line-height: 30px;"><span style="color: #222222;">dockboard</span> © Copyright 2014 . All Rights Reserved</td>
                      </tr>
                    </table>
                    <table border="0" width="10" align="left" cellpadding="0" cellspacing="0" style="border-collapse:collapse; mso-table-lspace:0pt; mso-table-rspace:0pt;">
                      <tr>
                        <td height="20" width="10" style="font-size: 20px; line-height: 20px;">&nbsp;</td>
                      </tr>
                    </table>
                    <table border="0" align="right" cellpadding="0" cellspacing="0" style="border-collapse:collapse; mso-table-lspace:0pt; mso-table-rspace:0pt;" class="container580">
                      <tr>
                        <td align="center" style="color: #adb3ba; font-size: 13px; font-family: 'Source Sans pro', Arial, sans-serif; line-height: 30px;"><a href="" style="color: #adb3ba; text-decoration: none;">Privacy Policy</a>&nbsp;&nbsp;<span style="font-weight: 700; color: #2c3e50;">/</span>&nbsp;&nbsp;<a href="" style="color: #adb3ba; text-decoration: none;">Terms of Use</a>&nbsp;&nbsp;<span style="font-weight: 700; color: #2c3e50;">/</span>&nbsp;&nbsp;<a href="" style="color: #adb3ba; text-decoration: none;">Contact</a></td>
                      </tr>
                    </table>
                  </td>
                </tr>
                <tr>
                  <td height="20" style="font-size: 20px; line-height: 20px;">&nbsp;</td>
                </tr>
              </table>
            </td>
          </tr>
        </table>
      </td>
    </tr>

    <tr>
      <td height="30" style="font-size: 30px; line-height: 30px;">&nbsp;</td>
    </tr>
    <!-- ======= end footer ======= -->
  <tr>
</table>    
`
