/*
 *  Document   : setting.js
 *  Author     : Meaglith Ma <genedna@gmail.com> @genedna
 *  Description:
 *
 */

'use strict';

//Auth Page Module
angular.module('dashboard', ['ngRoute', 'ngMessages', 'ngCookies', 'angular-growl', 'ui.codemirror', 'ui.bootstrap', 'ui.codemirror'])
//App Config
.config(['growlProvider', function(growlProvider){
  growlProvider.globalTimeToLive(6000);
}])
.controller('AddRepositoryCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $window) {
  $scope.editorOptions = {
    lineWrapping: true,
    lineNumbers: true,
    theme:"foo bar",
    indentWithTabs:true,
  };
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
.controller('PublicRepositoryCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $window) {
  $scope.repoTop = [];
  $scope.repoBottom = [];
  $scope.user = {};
  $http.get('/w1/repositories')
    .success(function(data, status, headers, config) {
      $scope.user = data;
      var repositories = $scope.user.repositoryobjects;
      var count = 0;
      for (var i = 0; i < repositories.length; i++) {
        if (repositories[i].privated) {
          continue;
        }
        if (repositories[i].starts == null) {
          repositories[i].totalStars = 0;
        } else {
          repositories[i].totalStars = repositories[i].starts.length;
        }
        if (count > 6) {
          $scope.repoBottom.push(repositories[i]);
          continue;
        }
        count++;
        $scope.repoTop.push(repositories[i]);
      }
    })
  .error(function(data, status, headers, config) {
    growl.error(data.message);
  });
}])
.controller('RepositoriesCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$window', '$log', function($scope, $cookies, $http, growl, $location, $timeout, $window, $log) {
  $scope.repoTop = [];
  $scope.repoBottom = [];
  $scope.repoBottomShow = [];
  $scope.user = {};

  $scope.bigTotalItems = 0;
  $scope.maxSize = 10;
  $scope.perPage = 5;
  $scope.bigCurrentPage = 1;
  $scope.pagingShow = false;
  $http.get('/w1/repositories')
    .success(function(data, status, headers, config) {
      $scope.user = data;
      var repositories = $scope.user.repositoryobjects;
      var conut = 0;
      for (var i = 0; i < repositories.length; i++) {
        if (repositories[i].starts == null) {
          repositories[i].totalStars = 0;
        } else {
          repositories[i].totalStars = repositories[i].starts.length;
        }
        if (i > 5) {
          $scope.repoBottom.push(repositories[i]);
          continue;
        }
        $scope.repoTop.push(repositories[i]);
      }
      $scope.bigTotalItems = $scope.repoBottom.length;
      if ($scope.repoBottom.length > 5) {
        $scope.pagingShow = true;
      }
      $scope.repoBottomShow = $scope.repoBottom.slice(0, $scope.perPage);
    })
  .error(function(data, status, headers, config) {
    growl.error(data.message);
  });

  $scope.pageChanged = function() {
    $scope.repoBottomShow = $scope.repoBottom.slice(($scope.bigCurrentPage - 1) * $scope.perPage, $scope.bigCurrentPage * $scope.perPage);
  };

}])
.controller('OrgRepositoriesCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$window', '$routeParams', function($scope, $cookies, $http, growl, $location, $timeout, $window, $routeParams) {
  $scope.repoTop = [];

  $http.get('/w1/organizations/' + $routeParams.orgUUID + '/repo')
    .success(function(data, status, headers, config) {
      var repositories = data;

      for (var i = 0; i < repositories.length; i++) {
        if (repositories[i].starts == null) {
          repositories[i].totalStars = 0;
        } else {
          repositories[i].totalStars = repositories[i].starts.length;
        }
        $scope.repoTop.push(repositories[i]);
      }
    })
  .error(function(data, status, headers, config) {
    growl.error(data.message);
  });
}])
.controller('PrivateRepositoryCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $window) {
  $scope.repoTop = [];
  $scope.repoBottom = [];
  $scope.user = {};
  $http.get('/w1/repositories')
    .success(function(data, status, headers, config) {
      $scope.user = data;
      var repositories = $scope.user.repositoryobjects;
      var conut = 0;
      for (var i = 0; i < repositories.length; i++) {
        if (!repositories[i].privated) {
          continue;
        }
        if (repositories[i].starts == null) {
          repositories[i].totalStars = 0;
        } else {
          repositories[i].totalStars = repositories[i].starts.length;
        }
        if (conut > 6) {
          $scope.repoBottom.push(repositories[i]);
          continue;
        }
        $scope.repoTop.push(repositories[i]);
        conut++;
      }
    })
  .error(function(data, status, headers, config) {
    growl.error(data.message);
  });
}])
.controller('StarRepositoryCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $window) {

}])
.controller('DockerfileCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $window) {
  $scope.privated = {};
  $scope.namespaces = [];
  $scope.repository = {};

  //init user data
  $scope.addPrivilege = false
    $http.get('/w1/namespace')
    .success(function(data, status, headers, config) {
      $scope.namespaces = data;
      $scope.repository.namespace = data[0];
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

}])
.controller('OrganizationAddCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $window) {
  $http.defaults.headers.post['X-XSRFToken'] = base64_decode($cookies._xsrf.split('|')[0]);

  $scope.submit = function() {
    $http.post('/w1/organization', $scope.organization)
      .success(function(data, status, headers, config) {
        growl.info(data.message);
        $timeout(function() {
          $window.location.href = '/setting#/' + $scope.organization.organization + '/team/add';
        }, 3000);
      })
      .error(function(data, status, headers, config) {
        growl.error(data.message);
      });
  }
}])
.controller('OrganizationListCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', '$window', function($scope, $cookies, $http, growl, $location, $timeout, $window) {
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
//routes
.config(function($routeProvider, $locationProvider) {
  $routeProvider
    .when('/', {
      templateUrl: '/static/views/dashboard/repositories.html',
      controller: 'RepositoriesCtrl'
    })
  .when('/repo', {
    templateUrl: '/static/views/dashboard/repositories.html',
    controller: 'RepositoriesCtrl'
  })
  .when('/repo/add', {
    templateUrl: '/static/views/dashboard/repositoryadd.html',
    controller: 'AddRepositoryCtrl'
  })
  .when('/repo/public', {
    templateUrl: '/static/views/dashboard/repopublic.html',
    controller: 'PublicRepositoryCtrl'
  })
  .when('/repo/private', {
    templateUrl: '/static/views/dashboard/repoprivate.html',
    controller: 'PrivateRepositoryCtrl'
  })
  .when('/repo/star', {
    templateUrl: '/static/views/dashboard/repostar.html',
    controller: 'StarRepositoryCtrl'
  })
  .when('/comments', {
    templateUrl: '/static/views/dashboard/comment.html',
    controller: 'CommentCtrl'
  })
  .when('/repo/dockerfile', {
    templateUrl: '/static/views/dashboard/dockerfile.html',
    controller: 'DockerfileCtrl'
  })
  .when('/org/add', {
    templateUrl: '/static/views/dashboard/organizationadd.html',
    controller: 'OrganizationAddCtrl'
  })
  .when('/org/:orgUUID/repo', {
    templateUrl: '/static/views/dashboard/organizationrepo.html',
    controller: 'OrgRepositoriesCtrl'
  });
})
.directive('namespaceValidator', [function() {
  var USERNAME_REGEXP = /^([a-z0-9_]{3,30})$/;

  return {
    require: 'ngModel',
    restrict: '',
    link: function(scope, element, attrs, ngModel) {
      ngModel.$validators.namespace = function(value) {
        return USERNAME_REGEXP.test(value) || value == "";
      }
    }
  };
}]);
