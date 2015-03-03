/*
 *  Document   : setting.js
 *  Author     : Meaglith Ma <genedna@gmail.com> @genedna
 *  Description:
 *
 */

'use strict';

//Auth Page Module
angular.module('setting', ['ngRoute', 'ngMessages', 'ngCookies', 'angular-growl', 'angularFileUpload'])
    .controller('AddRepositoryCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $window) {
        $scope.privated = {};
        $scope.namespaces = {};
        $scope.repository = {};
        $scope.namespaceObject = {};

        //init user data
        $scope.addPrivilege = false
        $http.get('/w1/namespaces')
            .success(function(data, status, headers, config) {
                $scope.namespaces = data;
                /* $scope.repository.namespace = data[0];*/
                $scope.namespaceObject = data[0];
            })
            .error(function(data, status, headers, config) {

            });
        $scope.privated = {};
        $scope.privated.values = [{
            code: 0,
            name: "Public"
        }, {
            code: 1,
            name: "Private"
        }];

        $scope.privated.selection = $scope.privated.values[0];

        //deal with create repository
        $scope.createRepo = function() {
            if ($scope.repoCreateForm.$valid) {
                if ($scope.privated.selection.code == 1) {
                    $scope.repository.privated = true;
                } else {
                    $scope.repository.privated = false;
                }

                $scope.repository.namespace = $scope.namespaceObject.namespace;
                $scope.repository.namespacetype = $scope.namespaceObject.namespacetype;

                $http.defaults.headers.post['X-XSRFToken'] = base64_decode($cookies._xsrf.split('|')[0]);
                $http.post('/w1/repository', $scope.repository)
                    .success(function(data, status, headers, config) {
                        $scope.addPrivilege = true;
                        growl.info(data.message);
                        $timeout(function() {
                            $window.location.href = '/dashboard';
                        }, 3000);
                    })
                    .error(function(data, status, headers, config) {
                        growl.error(data.message);
                    });
            }
        }

    }])
    .controller('SettingProfileCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', function($scope, $cookies, $http, growl, $location, $timeout, $upload) {
        //init user info
        $http.get('/w1/profile')
            .success(function(data, status, headers, config) {
                $scope.user = data
            })
            .error(function(data, status, headers, config) {

            });

        //deal with fileupload start
        $scope.submmitting = false;
        $scope.upload = function(files) {
            $http.defaults.headers.post['X-XSRFToken'] = base64_decode($cookies._xsrf.split('|')[0]);
            if (files && files.length) {
                for (var i = 0; i < files.length; i++) {
                    var file = files[i];
                    $upload.upload({
                        url: '/w1/gravatar',
                        file: file
                    }).progress(function(evt) {
                        //var progressPercentage = parseInt(100.0 * evt.loaded / evt.total);
                        //console.log('progress: ' + progressPercentage + '% ' +
                        //   evt.config.file.name);
                    }).success(function(data, status, headers, config) {
                        growl.info(data.message);
                        $scope.user.gravatar = data.url
                    });
                }
            }
        };

        $scope.submit = function() {
            if ($scope.profileForm.$valid) {
                $http.defaults.headers.put['X-XSRFToken'] = base64_decode($cookies._xsrf.split('|')[0]);
                $http.put('/w1/profile', $scope.user)
                    .success(function(data, status, headers, config) {
                        $scope.submmitting = true;
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
        $scope.submitting = false;
        $scope.submit = function() {
            if ($scope.accountForm.$valid) {
                $http.defaults.headers.put['X-XSRFToken'] = base64_decode($cookies._xsrf.split('|')[0]);
                $http.put('/w1/password', $scope.user)
                    .success(function(data, status, headers, config) {
                        $scope.submitting = true;
                        growl.info(data.message);
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
        $http.get('/w1/organizations/' + $routeParams.org)
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
        $scope.submitting = false;
        $scope.submit = function() {
            if (true) {
                $http.defaults.headers.put['X-XSRFToken'] = base64_decode($cookies._xsrf.split('|')[0]);
                $http.put('/w1/organization', $scope.organization)
                    .success(function(data, status, headers, config) {
                        $scope.submitting = true;
                        growl.info(data.message);
                    })
                    .error(function(data, status, headers, config) {
                        growl.error(data.message);
                    });
            }
        }
    }])
    .controller('SettingOrganizationAddCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window) {
        $scope.submitting = false;
        $scope.submit = function() {
            if ($scope.org.$valid) {
                $http.defaults.headers.post['X-XSRFToken'] = base64_decode($cookies._xsrf.split('|')[0]);
                $http.post('/w1/organization', $scope.organization)
                    .success(function(data, status, headers, config) {
                        $scope.submitting = true;
                        $window.location.href = '/setting';
                        growl.info(data.message);
                    })
                    .error(function(data, status, headers, config) {
                        $scope.submitting = false;
                        growl.error(data.message);
                    });
            }
        }

        $scope.createOrg = function() {
            $window.location.href = '/setting#/org/add';
        }
    }])
    .controller('SettingTeamCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', '$routeParams', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window, $routeParams) {
        //get teams data
        $scope.repositoryAdd = {};
        $scope.repo = {};
        $http.get('/w1/' + $routeParams.org + '/teams')
            .success(function(data, status, headers, config) {
                $scope.teams = data;
                /*  if length(data) == 0 {
                      $scope.organizationShow = false;
                      return
                  }
                  $scope.organizationShow = true;*/
            })
            .error(function(data, status, headers, config) {
                $timeout(function() {
                    //$window.location.href = '/auth';
                    alert(data.message);
                }, 3000);
            });

        $scope.createTeam = function() {
            $window.location.href = '/setting#/' + $routeParams.orgName + '/team/add';
        }

        $scope.editTeam = function(teamUUID) {
            $window.location.href = '/setting#/team/' + teamUUID + '/edit';
        }
    }])
    .controller('SettingTeamEditCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', '$routeParams', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window, $routeParams) {
        var availableTags = [];
        //get team data
        $http.get('/w1/team/' + $routeParams.teamUUID)
            .success(function(data, status, headers, config) {
                $scope.team = data;
                var usernames = [];
                for (var i = 0; i < $scope.team.userobjects.length; i++) {
                    usernames.push($scope.team.userobjects[i].username);
                }
                $scope.team.users = usernames;
            })
            .error(function(data, status, headers, config) {
                growl.error(data.message);
            });

        $http.get('/w1/users')
            .success(function(data, status, headers, config) {
                $scope.allUser = data;
                for (var i = 0; i < $scope.allUser.length; i++) {
                    availableTags.push($scope.allUser[i].username);
                }
            })
            .error(function(data, status, headers, config) {
                growl.error(data.message);
                return
            });
        $("#tags").autocomplete({
            source: availableTags
        });

        $scope.addUserFunc = function() {
            $scope.findUser = {}
            $scope.findUser.username = document.getElementById("tags").value;

            //adjust user already add
            for (var i = 0; i < $scope.team.userobjects.length; i++) {
                if ($scope.team.userobjects[i].username == document.getElementById("tags").value) {
                    growl.error("user already exist!");
                    $('#myModal').modal('toggle');
                    return;
                }
            }
            $scope.team.users.push($scope.findUser.username);
            $scope.team.userobjects.push($scope.findUser);
            $('#myModal').modal('toggle');
            return
        }

        $scope.removeUser = function(username) {
            var userUUIDSNew = [];
            var usersNew = [];
            for (var i = 0; i < $scope.team.userobjects.length; i++) {
                if ($scope.team.userobjects[i].username == username) {
                    continue;
                }
                userUUIDSNew.push($scope.team.userobjects[i].username);
                usersNew.push($scope.team.userobjects[i]);
            }
            $scope.team.users = userUUIDSNew;
            $scope.team.userobjects = usersNew;
            return
        }

        $scope.submit = function() {

            if ($scope.team.users.length == 0) {
                growl.error("Team must have members!");
                return
            }

            $http.defaults.headers.put['X-XSRFToken'] = base64_decode($cookies._xsrf.split('|')[0]);
            $http.put('/w1/team/' + $routeParams.teamUUID, $scope.team)
                .success(function(data, status, headers, config) {
                    $scope.submitting = true;
                    growl.info(data.message);
                })
                .error(function(data, status, headers, config) {
                    $scope.submitting = false;
                    growl.error(data.message);
                });

        }
    }])
    .controller('SettingTeamAddCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', '$routeParams', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window, $routeParams) {
        $scope.submitting = false;
        $scope.users = [];
        $scope.team = new Object();
        $scope.team.organization = $routeParams.org;
        $scope.team.users = [];
        var availableTags = [];
        //初始化organization数据
        $http.get('/w1/organizations')
            .success(function(data, status, headers, config) {
                $scope.organizations = data;
            })
            .error(function(data, status, headers, config) {
                growl.error(data.message);
                return
            });

        $http.get('/w1/users')
            .success(function(data, status, headers, config) {
                $scope.allUser = data;
                for (var i = 0; i < $scope.allUser.length; i++) {
                    availableTags.push($scope.allUser[i].username);
                }
            })
            .error(function(data, status, headers, config) {
                growl.error(data.message);
                return
            });
        $("#tags").autocomplete({
            source: availableTags
        });

        $scope.submit = function() {
            if ($scope.teamForm.$valid) {
                if ($scope.team.users.length == 0) {
                    growl.error("Team must have members!");
                    return
                }
                $http.defaults.headers.post['X-XSRFToken'] = base64_decode($cookies._xsrf.split('|')[0]);
                $http.post('/w1/team', $scope.team)
                    .success(function(data, status, headers, config) {
                        $scope.submitting = true;
                        growl.info(data.message);
                    })
                    .error(function(data, status, headers, config) {
                        $scope.submitting = false;
                        growl.error(data.message);
                    });
            }
        }


        $scope.addUserFunc = function() {
            $scope.findUser = {}
            $scope.findUser.username = document.getElementById("tags").value;

            //adjust user already add
            for (var i = 0; i < $scope.users.length; i++) {
                if ($scope.users[i].username == document.getElementById("tags").value) {
                    growl.error("user already exist!");
                    $('#myModal').modal('toggle');
                    return;
                }
            }
            $scope.users.push($scope.findUser);
            $scope.team.users.push($scope.findUser.username);
            $('#myModal').modal('toggle');
        }

        $scope.removeUser = function(username) {
            var newUsers = [];
            var newUsernames = [];
            for (var i = 0; i < $scope.users.length; i++) {
                if ($scope.users[i].username == username) {
                    continue;
                }
                newUsers.push($scope.users[i]);
                newUsernames.push($scope.users[i].username);
            }
            $scope.users = newUsers;
            $scope.team.users = newUsernames;
        }

    }])
    .controller('OrganizationListCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window) {
        //init organization  info
        $http.get('/w1/organizations')
            .success(function(data, status, headers, config) {
                $scope.organizaitons = data;
            })
            .error(function(data, status, headers, config) {
                $timeout(function() {
                    //$window.location.href = '/auth';

                }, 3000);
            });
    }])
    .controller('SettingCompetenceCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', '$routeParams', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window, $routeParams) {
        //get teams data
        $scope.repositoryAdd = {};
        $scope.repo = {};
        $http.get('/w1/' + $routeParams.org + '/teams')
            .success(function(data, status, headers, config) {
                $scope.teams = data;
            })
            .error(function(data, status, headers, config) {
                $timeout(function() {
                    //$window.location.href = '/auth';
                    alert(data.message);
                }, 3000);
            });
        //get repotories data
        $http.get('/w1/organizations/' + $routeParams.org + '/repo')
            .success(function(data, status, headers, config) {
                $scope.organiztionrepos = data;
                $scope.repositoryAdd = data[0];
            })
            .error(function(data, status, headers, config) {
                $timeout(function() {
                    //$window.location.href = '/auth';
                    alert(data.message);
                }, 3000);
            });

        //
        $scope.repo.privilege = "false";

        $scope.open = function(teamUUID) {
            $('#myModal').modal('toggle');
            $scope.repo.teamUUID = teamUUID;
        }

        $scope.addRepo4Team = function() {
            $scope.repo.repoUUID = $scope.repositoryAdd.UUID;
            $http.defaults.headers.post['X-XSRFToken'] = base64_decode($cookies._xsrf.split('|')[0]);
            $http.post('/w1/team/privilege', $scope.repo)
                .success(function(data, status, headers, config) {
                    $('#myModal').modal('toggle');
                    growl.info(data.message);
                    $http.get('/w1/' + $routeParams.org + '/teams')
                        .success(function(data, status, headers, config) {
                            $scope.teams = data;
                        })
                        .error(function(data, status, headers, config) {
                            $timeout(function() {
                                //$window.location.href = '/auth';
                                alert(data.message);
                            }, 3000);
                        });
                })
                .error(function(data, status, headers, config) {
                    $('#myModal').modal('toggle');
                    growl.error(data.message);
                });
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
            .when('/repo/add', {
                templateUrl: '/static/views/dashboard/repositoryadd.html',
                controller: 'AddRepositoryCtrl'
            })
            .when('/org/add', {
                templateUrl: '/static/views/setting/organizationadd.html',
                controller: 'SettingOrganizationAddCtrl'
            })
            .when('/org/:org', {
                templateUrl: '/static/views/setting/organization.html',
                controller: 'SettingOrganizationCtrl'
            })
            .when('/:org/team/add', {
                templateUrl: '/static/views/setting/teamadd.html',
                controller: 'SettingTeamAddCtrl'
            })
            .when('/team/:teamUUID/edit', {
                templateUrl: '/static/views/setting/teamedit.html',
                controller: 'SettingTeamEditCtrl'
            })
            .when('/:org/team/:orgName', {
                templateUrl: '/static/views/setting/team.html',
                controller: 'SettingTeamCtrl'
            })
            .when('/:org/competence', {
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
        var USERNAME_REGEXP = /^([a-z0-9_]{6,30})$/;

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