/*
 * wir-panel source code
 */

(function(){

function login(e) {
	e.preventDefault();

	this.loginInfo = {
		addr: $("#login-addr").val(),
		port: $("#login-port").val(),
		user: $("#login-username").val(),
		pass: $("#login-password").val()
	};

	$("#login").addClass("hidden");
	$("#remote").removeClass("hidden");
	$("#images").removeClass("hidden");
	$("#machines").removeClass("hidden");
}

var vm = new Vue({
	el: "#app",

	data: {
		remote: {
			addr: "127.0.0.1",
			port: 1997,
			cpu: 42,
			ramUsed: 1.2,
			ramTotal: 16,
			diskFree: 1642,
			backends: ["lxc", "qemu", "openvz"]
		},

		images: [
			{
				name: "qemu-alpine",
				type: "qemu",
				distro: "alpine linux",
				arch: "amd64",
				release: "3.3"
			},
			{
				name: "lxc-alpine",
				type: "lxc",
				distro: "alpine linux",
				arch: "amd64",
				release: "3.3"
			}
		],

		machines: [
			{
				name: "1fb8b95d1fe65c14",
				image: "qemu-alpine",
				state: "running",
				stats: {
					cpu: 42,
					ramUsed: 0.86,
					ramTotal: 2,
					diskUsed: 4.6,
					diskTotal: 45
				}
			},
			{
				name: "582838c01be65c14",
				image: "lxc-alpine",
				state: "running",
				stats: {
					cpu: 8,
					ramUsed: 0.16,
					ramTotal: 2,
					diskUsed: 1.1,
					diskTotal: 20
				}
			}
		]
	},

	methods: {
		login: login
	}
});

})();
