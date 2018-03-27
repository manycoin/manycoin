Pod::Spec.new do |spec|
  spec.name         = 'Gokc'
  spec.version      = '{{.Version}}'
  spec.license      = { :type => 'GNU Lesser General Public License, Version 3.0' }
  spec.homepage     = 'https://github.com/okcoin/go-okcoin'
  spec.authors      = { {{range .Contributors}}
		'{{.Name}}' => '{{.Email}}',{{end}}
	}
  spec.summary      = 'iOS Okcoin Client'
  spec.source       = { :git => 'https://github.com/okcoin/go-okcoin.git', :commit => '{{.Commit}}' }

	spec.platform = :ios
  spec.ios.deployment_target  = '9.0'
	spec.ios.vendored_frameworks = 'Frameworks/Gokc.framework'

	spec.prepare_command = <<-CMD
    curl https://gokcstore.blob.core.windows.net/builds/{{.Archive}}.tar.gz | tar -xvz
    mkdir Frameworks
    mv {{.Archive}}/Gokc.framework Frameworks
    rm -rf {{.Archive}}
  CMD
end
