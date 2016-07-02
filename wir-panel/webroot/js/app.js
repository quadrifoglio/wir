/*
 * wir-panel source code
 */

(function(){

var LoginInfo = null;

function login(e) {
	e.preventDefault();

	LoginInfo = {
		url: "http://" + $("#login-addr").val() + ":" + $("#login-port").val(),
		user: $("#login-username").val(),
		pass: $("#login-password").val()
	};

	$("#login").addClass("hidden");
	$("#remote").removeClass("hidden");
	$("#images").removeClass("hidden");
	$("#machines").removeClass("hidden");

	init(this);
}

function init(self) {
	$.ajax({method: "GET", url: LoginInfo.url + "/", success: function(res) {
		var info = null;

		if(info = JSON.parse(res)) {
			self.$data.remote.cpu = info.Content.Stats.CPUUsage;
			self.$data.remote.ramUsed = info.Content.Stats.RAMUsage;
			self.$data.remote.ramTotal = info.Content.Stats.RAMTotal;
			self.$data.remote.diskFree = info.Content.Stats.FreeSpace;
		}
		else {
			console.log("error");
		}
	}, error: function(res) {
		console.log("error");
	}});

	$.ajax({method: "GET", url: LoginInfo.url + "/images", success: function(res) {
		var imgs = null;

		if(imgs = JSON.parse(res)) {
			imgs.Content.forEach(function(i) {
				if(!i.Distro) i.Distro = "unknown os";
				if(!i.Arch) i.Arch = "unknown architecture";
			});
			self.$data.images = imgs.Content;
		}
		else {
			console.log("error");
		}
	}, error: function(res) {
		console.log("error");
	}});

	$.ajax({method: "GET", url: LoginInfo.url + "/machines", success: function(res) {
		var machines = null;

		if(machines = JSON.parse(res)) {
			machines.Content.forEach(function(m) {
				if(m.State == 1)
					m.State = "running";
				else
					m.State = "stopped";
			});

			self.$data.machines = machines.Content;
		}
		else {
			console.log("error");
		}
	}, error: function(res) {
		console.log("error");
	}});
}

var vm = new Vue({
	el: "#app",

	data: {
		remote: {
			addr: "",
			port: 0,
			cpu: 0,
			ramUsed: 0,
			ramTotal: 0,
			diskFree: 1642,
			backends: ["lxc", "qemu", "openvz"]
		},

		images: [],
		machines: []
	},

	methods: {
		login: login,
	}
});

})();
