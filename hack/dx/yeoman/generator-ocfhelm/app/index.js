'use strict';

const Generator = require('yeoman-generator');
const mkdirp = require('mkdirp');
const path = require('path');

module.exports = class extends Generator {
  constructor(args, opts) {
    super(args, opts);
    this.ucfirst = function(str) {
      var firstLetter = str.slice(0,1);
      return firstLetter.toUpperCase() + str.substring(1);
    };
  }

  async prompting() {
      this.log('\n' +
        '+-+-+-+-+-+-+-+-+-+-\n' +
        '|OCF Helm Generator|\n' +
        '+-+-+-+-+-+-+-+-+-+-\n' +
        '\n'
      );

      this.answers = await this.prompt([{
        type: 'input',
        name: 'categoryName',
        message: 'What is the category for this application (database, productivity, etc) [productivity]?',
        store   : true,
        default: 'productivity'
      },
      {
        type: 'input',
        name: 'vendorName',
        message: 'What is the vendor for the application\'s implementation (atlassian, gitlab, etc) [atlassian]?',
        store   : true,
        default: 'atlassian'
      },
      {
        type: 'input',
        name: 'baseName',
        message: 'What is the name of your application (jira, confluence, crowd, etc)? [confluence]',
        store   : true,
        default: 'confluence'
      },
      {
        type: 'input',
        name: 'interfaceDescription',
        message: 'What is the description of the interface? [Confluence is your remote-friendly team workspace where knowledge and collaboration meet]',
        store   : true,
        default: 'Confluence is your remote-friendly team workspace where knowledge and collaboration meet'
      },
      {
        type: 'input',
        name: 'iconURL',
        message: 'What is the URL of the icon? [https://www.atlassian.com/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Confluence%20Software@2x-blue.png]',
        store   : true,
        default: 'https://www.atlassian.com/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Confluence%20Software@2x-blue.png'
      },
      {
        type: 'input',
        name: 'documentationURL',
        message: 'What is the URL of the application documentation? [https://support.atlassian.com/confluence-cloud/resources/]',
        store   : true,
        default: 'https://support.atlassian.com/confluence-cloud/resources/'
      },
      {
        type: 'input',
        name: 'supportURL',
        message: 'What is the URL of the application\'s support site? [https://www.atlassian.com/software/confluence]',
        store   : true,
        default: 'https://www.atlassian.com/software/confluence'
      },
      {
        type: 'input',
        name: 'maintainerEmail',
        message: 'What is the maintainer\'s email address? [paul@structsure.co]',
        store   : true,
        default: 'paul@structsure.co'
      },
      {
        type: 'input',
        name: 'maintainerName',
        message: 'What is the maintainer\'s name? [Paul Pietkiewicz]',
        store   : true,
        default: 'Paul Pietkiewicz'
      },
      {
        type: 'input',
        name: 'maintainerURL',
        message: 'What is the maintainer\'s organization URL? [www.structsure.com]',
        store   : true,
        default: 'www.structsure.com'
      },
      {
        type: 'input',
        name: 'helmRepoURL',
        message: 'What is the URL of the helm repository containing the chart? [https://mox.sh/helm/]',
        store   : true,
        default: 'https://www.atlassian.com/software/confluence'
      },
      {
        type: 'input',
        name: 'helmChartName',
        message: 'What is the name of the helm chart? [confluence-server]',
        store   : true,
        default: 'confluence-server'
      },
      {
        type: 'input',
        name: 'helmChartDocumentationURL',
        message: 'What is the URL of the helm chart\'s documentation? [https://github.com/javimox/helm-charts/tree/master/charts/confluence-server]',
        store   : true,
        default: 'https://github.com/javimox/helm-charts/tree/master/charts/confluence-server'
      },
      {
        type: 'input',
        name: 'supportHelmURL',
        message: 'What is the URL of the helm chart\'s support site?  [https://mox.sh/helm]',
        store   : true,
        default: 'https://mox.sh/helm/'
      },
      {
        type: 'input',
        name: 'applicationVersion',
        message: 'What version of the application (in semver notation) [7.0.0]',
        store   : true,
        default: '7.0.0'
      }
    ]);
    }

  
  writing() {
      console.log ('Generating tree folders');

      const interfacesDir         = this.destinationPath('interfaces', this.answers.categoryName, this.answers.baseName),
            typesDir              = this.destinationPath('types', this.answers.categoryName, this.answers.baseName),
            implementationsDir    = this.destinationPath('implementations', this.answers.vendorName, this.answers.baseName);

      mkdirp.sync(implementationsDir);
      mkdirp.sync(interfacesDir);
      mkdirp.sync(typesDir);

      console.log ('Generating type yaml files');
      this.fs.copyTpl(
        this.templatePath('_type_install-input.yaml'),
        this.destinationPath(typesDir, 'install-input.yaml'),
        {
          categoryName: this.answers.categoryName,
          baseName: this.answers.baseName,
          capBaseName: this.ucfirst(this.answers.baseName),
          documentationURL: this.answers.documentationURL,
          supportURL: this.answers.supportURL,
          iconURL: this.answers.iconURL,
          maintainerEmail: this.answers.maintainerEmail,
          maintainerName: this.answers.maintainerName,
          maintainerURL: this.answers.maintainerURL
        }
      );
      this.fs.copyTpl(
        this.templatePath('_type_config.yaml'),
        this.destinationPath(typesDir, 'config.yaml'),
        {
          categoryName: this.answers.categoryName,
          baseName: this.answers.baseName,
          capBaseName: this.ucfirst(this.answers.baseName),
          documentationURL: this.answers.documentationURL,
          supportURL: this.answers.supportURL,
          iconURL: this.answers.iconURL,
          maintainerEmail: this.answers.maintainerEmail,
          maintainerName: this.answers.maintainerName,
          maintainerURL: this.answers.maintainerURL
        }
      );


      console.log ('Generating interface yaml files');
      this.fs.copyTpl(
        this.templatePath('_interface_PRODUCT.yaml'),
        this.destinationPath('interfaces', this.answers.categoryName, this.answers.baseName + '.yaml'),
        {
          categoryName: this.answers.categoryName,
          vendorName: this.answers.vendorName,
          baseName: this.answers.baseName,
          capBaseName: this.ucfirst(this.answers.baseName),
          iconURL: this.answers.iconURL,
          supportURL: this.answers.supportURL,
          interfaceDescription: this.answers.interfaceDescription,
          documentationURL: this.answers.documentationURL,
          maintainerEmail: this.answers.maintainerEmail,
          maintainerName: this.answers.maintainerName,
          maintainerURL: this.answers.maintainerURL,
        }
      );

      mkdirp.sync(interfacesDir + '/'+ this.answers.baseName);
      this.fs.copyTpl(
        this.templatePath('_interface_install.yaml'),
        this.destinationPath(interfacesDir, 'install.yaml'),
        {
          categoryName: this.answers.categoryName,
          vendorName: this.answers.vendorName,
          baseName: this.answers.baseName,
          capBaseName: this.ucfirst(this.answers.baseName),
          supportURL: this.answers.supportURL,
          iconURL: this.answers.iconURL,
          interfaceDescription: this.answers.interfaceDescription,
          documentationURL: this.answers.documentationURL,
          maintainerEmail: this.answers.maintainerEmail,
          maintainerName: this.answers.maintainerName,
          maintainerURL: this.answers.maintainerURL,
        }
      );
      this.fs.copyTpl(
        this.templatePath('_interface_uninstall.yaml'),
        this.destinationPath(interfacesDir, 'uninstall.yaml'),
        {
          categoryName: this.answers.categoryName,
          vendorName: this.answers.vendorName,
          baseName: this.answers.baseName,
          capBaseName: this.ucfirst(this.answers.baseName),
          supportURL: this.answers.supportURL,
          iconURL: this.answers.iconURL,
          interfaceDescription: this.answers.interfaceDescription,
          documentationURL: this.answers.documentationURL,
          maintainerEmail: this.answers.maintainerEmail,
          maintainerName: this.answers.maintainerName,
          maintainerURL: this.answers.maintainerURL,
        }
      );


      console.log ('Generating implementation installation yaml');
      this.fs.copyTpl(
        this.templatePath('_implementation_install.yaml'),
        this.destinationPath(implementationsDir,'install.yaml'),
        {
          categoryName: this.answers.categoryName,
          vendorName: this.answers.vendorName,
          baseName: this.answers.baseName,
          capBaseName: this.ucfirst(this.answers.baseName),
          documentationURL: this.answers.documentationURL,
          supportHelmURL: this.answers.supportHelmURL,
          maintainerEmail: this.answers.maintainerEmail,
          maintainerName: this.answers.maintainerName,
          maintainerURL: this.answers.maintainerURL,
          helmChartName: this.answers.helmChartName,
          helmRepoURL: this.answers.helmRepoURL
        }
      );
    }
  };
