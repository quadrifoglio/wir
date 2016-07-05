/*
 * wir-panel source code
 */

(function(){

var LoginInfo = null;

var VM = new Vue({
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

		errors: [],
		images: [],
		machines: []
	},

	methods: {
		login: login,
	}
});

function error(msg, fatal) {
	VM.$data.errors.push(msg);

	if(fatal) {
		$("#remote").addClass("hidden");
		$("#images").addClass("hidden");
		$("#machines").addClass("hidden");
	}
}

function login(e) {
	var self = this;

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

	$.ajax({method: "GET", url: LoginInfo.url + "/", success: function(res) {
		var info = null;

		if(info = JSON.parse(res)) {
			self.$data.remote.cpu = info.Content.Stats.CPUUsage;
			self.$data.remote.ramUsed = info.Content.Stats.RAMUsage;
			self.$data.remote.ramTotal = info.Content.Stats.RAMTotal;
			self.$data.remote.diskFree = info.Content.Stats.FreeSpace;

			listImages(self);
			listMachines(self);
		}
		else {
			error("Failed to parse output: " + res, true);
		}
	}, error: function(err) {
		error("Failed to connect: Failed to GET / : " + err.statusText, true);
	}});
}

function listImages(self) {
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
			error("Failed to parse output: " + res);
		}
	}, error: function(res) {
		error("Failed to GET /images : " + err.statusText);
	}});
}

function listMachines(self) {
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
			error("Failed to parse output: " + res);
		}
	}, error: function(res) {
		error("Failed to GET /machines : " + err.statusText);
	}});
}

})();
