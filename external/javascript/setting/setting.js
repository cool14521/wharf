/*
 *  Document   : setting.js
 *  Author     : Meaglith Ma <genedna@gmail.com> @genedna
 *  Description:
 *
 */

'use strict';

//Auth Page Module
angular.module('setting', ['ngRoute', 'ngMessages', 'ngCookies', 'angular-growl', 'angularFileUpload', 'ui.bootstrap'])
//App Config
.config(['growlProvider', function(growlProvider){
  growlProvider.globalTimeToLive(3000);
}])
//User Profile 
.controller('SettingProfileCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', function($scope, $cookies, $http, growl, $location, $timeout, $upload) {
  $scope.progress = 0;

  $http.get('/w1/profile')
    .success(function(data, status, headers, config) {
      $scope.user = data;
    })
  .error(function(data, status, headers, config) {
    growl.error(data.message);
  });

  //upload user profile    
  $scope.upload = function(files) {
    if (files && files.length) {
      for (var i = 0; i < files.length; i++) {
        $scope.progress = 0;
        var file = files[i];
        $upload.upload({
          url: '/w1/gravatar',
          file: file
        }).progress(function(evt) {
          $scope.progress = parseInt(100.0 * evt.loaded / evt.total);
        }).success(function(data, status, headers, config) {
          growl.info(data.message);
          $scope.user.gravatar = data.url
        });
      }
    }
  };

  //submit user profile
  $scope.submit = function() {
    if ($scope.profileForm.$valid) {
      $http.put('/w1/profile', $scope.user)
        .success(function(data, status, headers, config) {
          growl.info(data.message);
        })
        .error(function(data, status, headers, config) {
          growl.error(data.message);
        });
    }
  }
}])
//Reset Password
.controller('SettingAccountCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window) {
  $scope.submitting = false;
  $scope.submit = function() {
    if ($scope.accountForm.$valid) {
      $http.put('/w1/password', $scope.user)
        .success(function(data, status, headers, config) {
          $scope.submitting = true;
          growl.info(data.message);
          //Reset Input Emptry
          $scope.user.oldPassword = "";
          $scope.user.newPassword = "";
          $scope.user.password_confirm = "";
          //Clean Input Validate
          $scope.accountForm.password.$dirty = false;
          $scope.accountForm.newPassword.$dirty = false;
          $scope.accountForm.password_confirm.$dirty = false;
        })
      .error(function(data, status, headers, config) {
        $scope.submitting = false;
        growl.error(data.message);
      });
    }
  }
}])
//Email Setting
.controller('SettingEmailsCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window) {
  $scope.submit = function(){}
}])
//Notification Setting
.controller('SettingNotificationCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window) {
  $scope.submit = function(){}
}])
//Organization Edit
.controller('SettingOrganizationCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', '$routeParams', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window, $routeParams) {
  //Get Organization
  $http.get('/w1/organizations/' + $routeParams.org)
    .success(function(data, status, headers, config) {
      $scope.organization = data;
    })
    .error(function(data, status, headers, config) {
      growl.error(data.message);
    });

  //Submit Organization
  $scope.submit = function() {
    $http.put('/w1/organization', $scope.organization)
      .success(function(data, status, headers, config) {
        growl.info(data.message);
      })
      .error(function(data, status, headers, config) {
        growl.error(data.message);
      });
  }
}])
//Team List
.controller('SettingTeamListCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', '$routeParams', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window, $routeParams) {
  //Get teams and members in team
  $http.get('/w1/' + $routeParams.org + '/teams')
    .success(function(data, status, headers, config) {
      data.forEach(
        function getTeam(value) {
          value.profiles = [];
          value.users.forEach(
            function getUser(username){
              $http.get('/w1/profile/' + username)
                .success(function(profile, status, headers, config) {
                  value.profiles.push({url: '/u/' + profile.username, gravatar: profile.gravatar});
                })
                .error(function(profile, status, headers, config) {
                  growl.error(profile.message);
                });
            }
          );
        }
      );

      $scope.teams = data;
    })
    .error(function(data, status, headers, config) {
      growl.error(data.message);
    });

  //Create Team Button
  $scope.create = function() {
    $window.location.href = '/setting#/' + $routeParams.org + '/team/add';
  }

  //Edit Team Button
  $scope.edit = function(name) {
    $window.location.href = '/setting#/team/' + name + '/edit';
  }
}])
//Team Edit And Add & Remove Users
.controller('SettingTeamEditCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', '$routeParams', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window, $routeParams) {
  $scope.users = [];
  $scope.team = {};
  $scope.team.organization = $routeParams.org;

  //get team data
  $http.get('/w1/' + $routeParams.org + '/team/' + $routeParams.team)
    .success(function(data, status, headers, config) {
      $scope.team = data;
      for(var i = 0; i < data.users.length; i++){

        $http.get('/w1/profile/' + data.users[i])
          .success(function(profile, status, headers, config) {
            $scope.users.push({
              url: '/u/' + profile.username,
              username: profile.username,
              gravatar: profile.gravatar,
              name: profile.fullname
            });
          })
          .error(function(profile, status, headers, config) {
            growl.error(profile.message);
          });
      }
    })
    .error(function(data, status, headers, config) {
      growl.error(data.message);
    });

  //Get all users for typeahead
  $http.get('/w1/users')
    .success(function(data, status, headers, config) {
      var users = new Bloodhound({
        datumTokenizer: function(d) { return Bloodhound.tokenizers.whitespace(d.username); },
        queryTokenizer: Bloodhound.tokenizers.whitespace,
        limit: 10,
        local: data
      });

      users.initialize();

      $('input.typeahead').typeahead(null, {
        name: 'states',
        displayKey: 'username',
        source: users.ttAdapter()
      });
    })
    .error(function(data, status, headers, config) {
      growl.error(data.message);
    });

  //Add a user
  $scope.add = function() {
    var user = document.getElementsByName("finding")[0].value;

    if (user.length > 0){
      $http.put('/w1/' + $routeParams.org + '/team/' + $routeParams.team + '/add/' + user)
        .success(function(data, status, headers, config){
          $scope.users.push({
            url: '/u/' + data.username,
            username: data.username,
            gravatar: data.gravatar,
            name: data.fullname
          });
          document.getElementsByName("finding")[0].value = null;
          $('.typeahead').typeahead('val', '');
        })
        .error(function(data, status, headers, config){
          growl.error(data.message);
          document.getElementsByName("finding")[0].value = null;
          $('.typeahead').typeahead('val', '');
        });
    }
  }
  
  //Remove a user
  $scope.remove = function(user) {
    $http.put('/w1/' + $routeParams.org + '/team/' + $routeParams.team + '/remove/' + user)
      .success(function(data, status, headers, config){
        $scope.users = _.filter($scope.users, function(u){
          return u.username != user;
        });
      })
      .error(function(data, status, headers, config){
        growl.error(data.message);
      });
  }

  //Update a team
  $scope.submit = function() { 
    $http.put('/w1/team/' + $routeParams.team, $scope.team)
      .success(function(data, status, headers, config){
        growl.info(data.message);
      })
      .error(function(data, status, headers, config){
        growl.error(data.message);
      });  
  }

}])
.controller('SettingTeamAddCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', '$routeParams', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window, $routeParams) {
  $scope.team = {};
  $scope.team.organization = $routeParams.org;

  //Add a team
  $scope.submit = function() {
    if ($scope.teamForm.$valid){
      $http.post('/w1/team', $scope.team)
        .success(function(data, status, headers, config) {
          $scope.submitting = true;
          growl.info(data.message);
          $timeout(function() {
            $window.location.href = "/setting#/" + $routeParams.org + "/teams";
          }, 3000);
        })
        .error(function(data, status, headers, config) {
          $scope.submitting = false;
          growl.error(data.message);
        });
    }
  }
}])
.controller('OrganizationListCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window) {
  //init organization  info
  $http.get('/w1/organizations')
    .success(function(data, status, headers, config) {
      $scope.organizaitons = data;
    })
    .error(function(data, status, headers, config) {
      growl.error(data.message);
    });
}])
.controller('SettingTeamPermissionCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$upload', '$window', '$routeParams', function($scope, $cookies, $http, growl, $location, $timeout, $upload, $window, $routeParams) {
  
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
  .when('/:org/team/add', {
    templateUrl: '/static/views/setting/teamadd.html',
    controller: 'SettingTeamAddCtrl'
  })
  .when('/:org/team/:team', {
    templateUrl: '/static/views/setting/teamedit.html',
    controller: 'SettingTeamEditCtrl'
  })
  .when('/:org/teams', {
    templateUrl: '/static/views/setting/team.html',
    controller: 'SettingTeamListCtrl'
  })
  .when('/:org/permissions', {
    templateUrl: '/static/views/setting/competence.html',
    controller: 'SettingTeamPermissionCtrl'
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
  var USERNAME_REGEXP = /^([a-z0-9_]{3,30})$/;

  return {
    require: 'ngModel',
    restrict: '',
    link: function(scope, element, attrs, ngModel) {
      ngModel.$validators.usernames = function(value) {
        return USERNAME_REGEXP.test(value) || value == "";
      }
    }
  };
}])
.directive('ngEnter', function () {
  return function (scope, element, attrs) {
    element.bind("keydown keypress", function (event) {
      if(event.which === 13) {
        scope.$apply(function (){
          scope.$eval(attrs.ngEnter);
        });

        event.preventDefault();
      }
    });
  };
});;
