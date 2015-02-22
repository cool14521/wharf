/*
 *  Document   : setting.js
 *  Author     : Meaglith Ma <genedna@gmail.com> @genedna
 *  Description:
 *
 */

'use strict';

//Auth Page Module
angular.module('setting', ['ngRoute', 'ngMessages', 'ngCookies', 'angular-growl', 'angularFileUpload'])
    .controller('SettingProfileCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', function($scope, $cookies, $http, growl, $location, $timeout, $upload) {
        //init user info
        $http.get('/w1/profile')
            .success(function(data, status, headers, config) {
                $scope.user = data
            })
            .error(function(data, status, headers, config) {
                
            });

        //deal with fileupload start
        var version = '1.3.8';
        $scope.fileReaderSupported = window.FileReader != null && (window.FileAPI == null || FileAPI.html5 != false);
        $scope.changeAngularVersion = function() {
            window.location.hash = $scope.angularVersion;
            window.location.reload(true);
        };
        $scope.angularVersion = window.location.hash.length > 1 ? (window.location.hash.indexOf('/') === 1 ?
            window.location.hash.substring(2) : window.location.hash.substring(1)) : '1.2.20';
        // you can also $scope.$watch('files',...) instead of calling upload from ui
        $scope.upload = function(files) {
            $scope.formUpload = false;
            if (files != null) {
                for (var i = 0; i < files.length; i++) {
                    $scope.errorMsg = null;
                    (function(file) {
                        $scope.generateThumb(file);
                        eval($scope.uploadScript);
                    })(files[i]);
                }
            }
        };
        $scope.generateThumb = function(file) {
            if (file != null) {
                if ($scope.fileReaderSupported && file.type.indexOf('image') > -1) {
                    $timeout(function() {
                        var fileReader = new FileReader();
                        fileReader.readAsDataURL(file);
                        fileReader.onload = function(e) {
                            file.dataUrl = e.target.result;
                            $timeout(function() {
                                file.upload = $upload.upload({
                                    url: '/w1/gravatar', //upload.php script, node.js route, or servlet url
                                    method: 'POST',
                                    headers: {
                                        'Content-Type': "multipart/form-data",
                                        'X-XSRFToken': base64_decode($cookies._xsrf.split('|')[0])
                                    },
                                    data: {
                                        filename: 'file'
                                    },
                                    file: file // or list of files ($files) for html5 only
                                }).progress(function(evt) {
                                    file.progress = Math.min(100, parseInt(100.0 * evt.loaded / evt.total));
                                }).success(function(data, status, headers, config) { // file is uploaded successfully
                                    growl.info(data.message);
                                    $scope.user.gravatar = data.url
                                }).error(function(data, status, headers, config) {
                                    growl.error(data.message);
                                });
                            });
                        }
                    });
                }
            }
        }
        if (localStorage) {
            $scope.acl = localStorage.getItem("acl");
            $scope.success_action_redirect = localStorage.getItem("success_action_redirect");
            $scope.policy = localStorage.getItem("policy");
        }
        $scope.success_action_redirect = $scope.success_action_redirect || window.location.protocol + "//" + window.location.host;
        $scope.jsonPolicy = $scope.jsonPolicy || '{\n "expiration": "2020-01-01T00:00:00Z",\n "conditions": [\n {"bucket": "angular-file-upload"},\n ["starts-with", "$key", ""],\n {"acl": "private"},\n ["starts-with", "$Content-Type", ""],\n ["starts-with", "$filename", ""],\n ["content-length-range", 0, 524288000]\n ]\n}';
        $scope.acl = $scope.acl || 'private';

        $scope.confirm = function() {
            return confirm('Are you sure? Your local changes will be lost.');
        }
        $scope.getReqParams = function() {
                return $scope.generateErrorOnServer ? "?errorCode=" + $scope.serverErrorCode +
                    "&errorMessage=" + $scope.serverErrorMsg : "";
            }
            //deal with fileupload end
        $scope.submit = function() {
            if ($scope.profileForm.$valid) {
                $http.defaults.headers.put['X-XSRFToken'] = base64_decode($cookies._xsrf.split('|')[0]);
                $http.put('/w1/profile', $scope.user)
                    .success(function(data, status, headers, config) {
                        growl.info(data.message);
                    })
                    .error(function(data, status, headers, config) {
                        $scope.submitting = false;
                        growl.error(data.message);
                    });
            }
        }
    }])
    .controller('SettingAccountCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window) {
        $scope.submit = function() {
            if ($scope.accountForm.$valid) {
                $http.defaults.headers.put['X-XSRFToken'] = base64_decode($cookies._xsrf.split('|')[0]);
                $http.put('/w1/account', $scope.user)
                    .success(function(data, status, headers, config) {
                        growl.info(data.message);
                        $window.location.href = '/setting';
                    })
                    .error(function(data, status, headers, config) {
                        $scope.submitting = false;
                        growl.error(data.message);
                    });
            }
        }
    }])
    .controller('SettingEmailsCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window) {
        $scope.submit = function() {
            $http.defaults.headers.put['X-XSRFToken'] = base64_decode($cookies._xsrf.split('|')[0]);
        }
    }])
    .controller('SettingNotificationCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window) {
        $scope.submit = function() {

        }
    }])
    .controller('SettingOrganizationCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', '$routeParams', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window, $routeParams) {
        $http.get('/w1/organizations/' + $routeParams.orgName)
            .success(function(data, status, headers, config) {
                $scope.organization = data;
                /*  if length(data) == 0 {
                      $scope.organizationShow = false;
                      return
                  }
                  $scope.organizationShow = true;*/
            })
            .error(function(data, status, headers, config) {
                $timeout(function() {
                    //$window.location.href = '/auth';
                    alert(data);
                }, 3000);
            });

        $scope.submit = function() {
            if (true) {
                $http.defaults.headers.put['X-XSRFToken'] = base64_decode($cookies._xsrf.split('|')[0]);
                $http.put('/w1/organization', $scope.organization)
                    .success(function(data, status, headers, config) {
                        growl.info(data.message);
                    })
                    .error(function(data, status, headers, config) {
                        $scope.submitting = false;
                        growl.error(data.message);
                    });
            }
        }
    }])
    .controller('SettingOrganizationAddCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window) {
        $scope.submit = function() {
            if (true) {
                $http.defaults.headers.post['X-XSRFToken'] = base64_decode($cookies._xsrf.split('|')[0]);
                $http.post('/w1/organization', $scope.organization)
                    .success(function(data, status, headers, config) {
                        growl.info(data.message);
                    })
                    .error(function(data, status, headers, config) {
                        $scope.submitting = false;
                        growl.error(data.message);
                    });
            }
        }

        $scope.createOrg = function() {
            $window.location.href = '/setting#/organizationAdd';
        }
    }])
    .controller('SettingTeamCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window) {
        //初始化加载user的organization信息
        $http.get('/w1/organization')
            .success(function(data, status, headers, config) {
                $scope.team = data
            })
            .error(function(data, status, headers, config) {

            });

        $scope.submit = function() {

        }

        $scope.createTeam = function() {
            $window.location.href = '/setting#/teamAdd';
        }
    }])
    .controller('SettingTeamAddCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window) {
        $scope.users = [];
        $scope.team = new Object();
        $scope.team.users = [];
        $scope.findUser = new Object();
        //初始化organization数据
        $http.get('/w1/organizations')
            .success(function(data, status, headers, config) {
                $scope.organizations = data;

            })
            .error(function(data, status, headers, config) {
                $timeout(function() {
                    //$window.location.href = '/auth';
                    alert(data);
                }, 3000);
            });

        $scope.submit = function() {
            if (true) {
                $http.defaults.headers.post['X-XSRFToken'] = base64_decode($cookies._xsrf.split('|')[0]);
                $http.post('/w1/team', $scope.team)
                    .success(function(data, status, headers, config) {
                        growl.info(data.message);
                    })
                    .error(function(data, status, headers, config) {
                        $scope.submitting = false;
                        growl.error(data.message);
                    });
            }
        }

        $scope.Search = function() {
            /*
            $http.get('/w1/users/'+$scope.user.username)
                .success(function(data, status, headers, config) {
                    $scope.users = data;

                })
                .error(function(data, status, headers, config) {
                    $timeout(function() {
                        //$window.location.href = '/auth';
                        alert(data);
                    }, 3000);
                });
            $('.dropdown-toggle').dropdown();
            */
        }

        var availableTags = ["chliang2030598"];

        $("#tags").autocomplete({
            source: availableTags
        });


        $scope.addUserFunc = function() {
            $scope.findUser.username = document.getElementById("tags").value;
            $scope.users.push($scope.findUser);
            $scope.team.users.push($scope.findUser.username);
            $('#myModal').modal('hide');
        }

    }])
    .controller('OrganizationListCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window) {
        //init organization  info
        $http.get('/w1/organizations')
            .success(function(data, status, headers, config) {
                $scope.organizaitons = data;
                /*  if length(data) == 0 {
                      $scope.organizationShow = false;
                      return
                  }
                  $scope.organizationShow = true;*/
            })
            .error(function(data, status, headers, config) {
                $timeout(function() {
                    //$window.location.href = '/auth';
                    alert(data);
                }, 3000);
            });
    }])
    .controller('SettingCompetenceCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window) {
        $scope.submit = function() {

        }
    }])
    //routes
    .config(function($routeProvider, $locationProvider) {
        $routeProvider
            .when('/', {
                templateUrl: '/static/views/setting/profile.html',
                controller: 'SettingProfileCtrl'
            })
            .when('/profile', {
                templateUrl: '/static/views/setting/profile.html',
                controller: 'SettingProfileCtrl'
            })
            .when('/account', {
                templateUrl: '/static/views/setting/account.html',
                controller: 'SettingAccountCtrl'
            })
            .when('/emails', {
                templateUrl: '/static/views/setting/emails.html',
                controller: 'SettingEmailsCtrl'
            })
            .when('/notification', {
                templateUrl: '/static/views/setting/notification.html',
                controller: 'SettingNotificationCtrl'
            })
            .when('/org/:org', {
                templateUrl: '/static/views/setting/organization.html',
                controller: 'SettingOrganizationCtrl'
            })
            .when('/org/add', {
                templateUrl: '/static/views/setting/organizationAdd.html',
                controller: 'SettingOrganizationAddCtrl'
            })
            .when('/team', {
                templateUrl: '/static/views/setting/team.html',
                controller: 'SettingTeamCtrl'
            })
            .when('/team/add', {
                templateUrl: '/static/views/setting/teamAdd.html',
                controller: 'SettingTeamAddCtrl'
            })
            .when('/competence', {
                templateUrl: '/static/views/setting/competence.html',
                controller: 'SettingCompetenceCtrl'
            });
    })
    .directive('emailValidator', [function() {
        var EMAIL_REGEXP = /^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;

        return {
            require: 'ngModel',
            restrict: '',
            link: function(scope, element, attrs, ngModel) {
                ngModel.$validators.emails = function(value) {
                    return EMAIL_REGEXP.test(value);
                }
            }
        };
    }])
    .directive('urlValidator', [function() {
        var URL_REGEXP = /(http|https):\/\/[\w\-_]+(\.[\w\-_]+)+([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?/;

        return {
            require: 'ngModel',
            restrict: '',
            link: function(scope, element, attrs, ngModel) {
                ngModel.$validators.urls = function(value) {
                    return URL_REGEXP.test(value) || value == "";
                }
            }
        };
    }])
    .directive('mobileValidator', [function() {
        var MOBILE_REGEXP = /^[0-9]{1,20}$/;

        return {
            require: 'ngModel',
            restrict: '',
            link: function(scope, element, attrs, ngModel) {
                ngModel.$validators.mobiles = function(value) {
                    return MOBILE_REGEXP.test(value) || value == "";
                }
            }
        };
    }])
    .directive('passwdValidator', [function() {
        var NUMBER_REGEXP = /\d/;
        var LETTER_REGEXP = /[A-z]/;

        return {
            require: 'ngModel',
            restrict: '',
            link: function(scope, element, attrs, ngModel) {
                ngModel.$validators.passwd = function(value) {
                    return NUMBER_REGEXP.test(value) && LETTER_REGEXP.test(value);
                }
            }
        };
    }])
    .directive('confirmValidator', [function() {
        return {
            require: 'ngModel',
            restrict: '',
            scope: {
                passwd: "=confirmData"
            },
            link: function(scope, element, attrs, ngModel) {
                ngModel.$validators.repeat = function(value) {
                    return value == scope.passwd;
                };
                scope.$watch('passwd', function() {
                    ngModel.$validate();
                });
            }
        };
    }])
    .directive('namespaceValidator', [function() {
        var NAMESPACE_REGEXP = /^([a-z0-9_]{6,30})$/;

        return {
            require: 'ngModel',
            restrict: '',
            link: function(scope, element, attrs, ngModel) {
                ngModel.$validators.usernames = function(value) {
                    return USERNAME_REGEXP.test(value) || value == "";
                }
            }
        };
    }]);
